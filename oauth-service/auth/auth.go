package auth

import (
	"errors"
	"fmt"
	"net/http"
	nethttputil "net/http/httputil"

	"github.com/darren-west/app/oauth-service/httputil"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

//go:generate mockgen -destination ./mocks/mock_store.go -package mocks github.com/gorilla/sessions Store

func WithLoginPattern(pattern string) Option {
	return func(options *Options) (err error) {
		if pattern == "" {
			err = fmt.Errorf("invalid option: login pattern is empty")
			return
		}
		options.LoginPattern = pattern
		return
	}
}

func WithRedirectPattern(pattern string) Option {
	return func(options *Options) (err error) {
		if pattern == "" {
			err = fmt.Errorf("invalid option: redirect pattern is empty")
			return
		}
		options.RedirectPattern = pattern
		return
	}
}

// WithSessionStore sets the backend session to use. This will be used to store the
// oauth2 state.
func WithSessionStore(store sessions.Store) Option {
	return func(opts *Options) (err error) {
		if store == nil {
			err = fmt.Errorf("invalid option: store is nil")
			return
		}
		opts.store = store
		return
	}
}

// WithOauth2Config sets the oauth2 config to use to authenticate.
func WithOauth2Config(config *oauth2.Config) Option {
	return func(opts *Options) (err error) {
		if config == nil {
			err = fmt.Errorf("invalid option: oauth config is nil")
			return
		}
		opts.Config = config
		return
	}
}

// WithOnAuthenticatedCb sets the callback to use when the user has been authenticated. It provides
// a ready to use http client with the token set.
func WithAuthenticator(a Authenticator) Option {
	return func(opts *Options) (err error) {
		if a == nil {
			err = fmt.Errorf("invalid option: authenticator is nil")
			return
		}
		opts.Authenticator = a
		return
	}
}

// Options are the handlers options. It is a struct for holding
// setable configuration.
type (
	Options struct {
		LoginPattern    string
		RedirectPattern string
		store           sessions.Store
		SessionName     string
		Config          *oauth2.Config
		Authenticator   Authenticator
	}

	// Option is used to set an option.
	Option func(*Options) error
)

// NewHandler returns a new handler or an error if any option is not valid.
func NewHandler(opts ...Option) (h Handler, err error) {
	h = newDefaultHandler()
	for _, opt := range opts {
		if err = opt(h.options); err != nil {
			return
		}
	}
	h.mux = http.NewServeMux()
	h.mux.HandleFunc(h.options.LoginPattern, h.login)
	h.mux.HandleFunc(h.options.RedirectPattern, h.loginRedirect)
	return
}

func newDefaultHandler() (h Handler) {
	return Handler{
		options: &Options{
			store:           sessions.NewCookieStore([]byte("abcd")),
			SessionName:     "auth",
			LoginPattern:    "/login",
			RedirectPattern: "/redirect",
		},
	}
}

// Handler is the type used to handle the login and Redirect http.Handles.
// DO NOT instantiate without using NewHandler().
type Handler struct {
	options *Options
	mux     *http.ServeMux
}

// login this handles a users login with the OAuth2 config passed into the NewHandler function. It will redirect
// the user to the OAuth2 login and handle updating the session.
func (h Handler) login(w http.ResponseWriter, r *http.Request) {
	h.do(w, r, func(ext httpExtension, w http.ResponseWriter, r *http.Request) (httpErr *httputil.Error) {
		stateValue := uuid.New().String()
		ext.Session.Values["state"] = stateValue
		if err := ext.Session.Save(r, w); err != nil {
			return httputil.NewError(http.StatusInternalServerError, fmt.Errorf("unable to save session: %s", err))
		}
		http.Redirect(w, r, h.options.Config.AuthCodeURL(stateValue), http.StatusFound)
		return
	})
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

//go:generate mockgen -destination ./mocks/mock_authenticator.go -package mocks github.com/darren-west/app/oauth-service/auth Authenticator
type (
	// UserInfo is a struct containing the authenticated user from the oauth server.
	UserInfo struct {
		ID        string
		FirstName string
		LastName  string
		Email     string
	}

	// Authenticator is an interface to allow different backend api retrieval of a user.
	Authenticator interface {
		RetrieveUser(*http.Client) (UserInfo, error)
		OnAuthenticated(http.ResponseWriter, UserInfo)
	}
)

//// OnAuthenticated is a callback function that is triggered when the user has been authenticated.
//// The client will be instantiated with the token and ready to use to hit an endpoint API.
//type OnAuthenticated func(*http.Client, http.ResponseWriter)

// loginRedirect handles the redirection call from the OAuth2 server. It will trigger the OnAuthenticated callback.
func (h Handler) loginRedirect(w http.ResponseWriter, r *http.Request) {
	h.do(w, r, func(ext httpExtension, w http.ResponseWriter, r *http.Request) (httpErr *httputil.Error) {
		if ext.Session.Values["state"] != r.URL.Query().Get("state") {
			return httputil.NewError(http.StatusUnauthorized, errors.New("state token invalid"))
		}
		token, err := h.options.Config.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
		if err != nil {
			return httputil.NewError(http.StatusInternalServerError, err)
		}
		user, err := h.options.Authenticator.RetrieveUser(h.options.Config.Client(oauth2.NoContext, token))
		if err != nil {
			return httputil.NewError(http.StatusInternalServerError, err)
		}
		h.options.Authenticator.OnAuthenticated(w, user)
		return
	})
}

type httpExtension struct {
	Session *sessions.Session
	Logger  *logrus.Entry
}

func (h Handler) do(w http.ResponseWriter, r *http.Request, f func(httpExtension, http.ResponseWriter, *http.Request) *httputil.Error) {
	session, err := h.session(r)
	if err != nil {
		httputil.NewError(http.StatusInternalServerError, fmt.Errorf("unable to read session: %s", err)).Write(w)
		return
	}
	ext := httpExtension{
		Logger:  newLoggerWithRequest(r),
		Session: session,
	}
	ext.Logger.Info("Received request.")
	if httpError := f(ext, w, r); httpError != nil {
		ext.Logger.WithError(err).Errorf("Failed to process http request.")
		httpError.Write(w)
	}
}

func (h Handler) session(r *http.Request) (sess *sessions.Session, err error) {
	if sess, err = h.options.store.Get(r, h.options.SessionName); err != nil {
		return
	}
	return
}

// Options returns the handlers Options.
func (h Handler) Options() Options {
	return *h.options
}

func newLoggerWithRequest(r *http.Request) *logrus.Entry {
	data, _ := nethttputil.DumpRequest(r, true)
	return logrus.WithField("http request", string(data))
}
