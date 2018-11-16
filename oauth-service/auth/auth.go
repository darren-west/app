package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	nethttputil "net/http/httputil"

	"github.com/darren-west/app/oauth-service/config"
	"github.com/darren-west/app/oauth-service/httputil"
	"github.com/darren-west/app/utils/session"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

//go:generate mockgen -destination ./mocks/mock_store.go -package mocks github.com/gorilla/sessions Store

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

// WithConfig sets the oauth2 config to use to authenticate.
func WithConfig(config config.Options) Option {
	return func(opts *Options) (err error) {
		opts.Config = config
		return
	}
}

// WithLoginHandler sets the login handler.
func WithLoginHandler(h LoginHandler) Option {
	return func(opts *Options) (err error) {
		opts.LoginHandler = h
		return
	}
}

// Options are the handlers options. It is a struct for holding
// setable configuration.
type (
	Options struct {
		store        sessions.Store
		Config       config.Options
		LoginHandler LoginHandler
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
	h.mux.HandleFunc(h.options.Config.LoginRoutePath, h.login)
	h.mux.HandleFunc(h.options.Config.RedirectRoutePath, h.redirect)
	return
}

func newDefaultHandler() (h Handler) {
	return Handler{
		options: &Options{
			store: sessions.NewCookieStore([]byte("abcd")),
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
		http.Redirect(w, r, h.options.Config.OAuth.AuthCodeURL(stateValue), http.StatusFound)
		return
	})
}

// ServeHTTP implements the http handler interface forwarding requests to the underlying handlers.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

type (
	// UserInfo is a struct containing the authenticated user from the oauth server.
	UserInfo struct {
		ID        string
		FirstName string
		LastName  string
		Email     string
	}
)

//go:generate mockgen -destination ./mocks/login_handler.go -package mocks github.com/darren-west/app/oauth-service/auth LoginHandler

// LoginHandler passes a user allowing custom logic to handle a user logged in by oauth.
type LoginHandler interface {
	Handle(UserInfo, http.ResponseWriter, *http.Request)
}

// loginRedirect handles the redirection call from the OAuth2 server. It will trigger the OnAuthenticated callback.
func (h Handler) redirect(w http.ResponseWriter, r *http.Request) {
	h.do(w, r, func(ext httpExtension, w http.ResponseWriter, r *http.Request) (httpErr *httputil.Error) {
		if ext.Session.Values["state"] != r.URL.Query().Get("state") {
			return httputil.NewError(http.StatusUnauthorized, errors.New("state token invalid"))
		}
		token, err := h.options.Config.OAuth.Exchange(oauth2.NoContext, r.URL.Query().Get("code"))
		if err != nil {
			return httputil.NewError(http.StatusInternalServerError, err)
		}
		client := h.options.Config.OAuth.Client(oauth2.NoContext, token)
		resp, err := client.Get(h.options.Config.APIEndpoint)
		if err != nil {
			return httputil.NewError(http.StatusInternalServerError, err)
		}
		user, err := decodeUser(resp.Body, h.options.Config.UserMapping)
		if err != nil {
			return httputil.NewError(http.StatusInternalServerError, err)
		}
		if h.options.LoginHandler != nil {
			h.options.LoginHandler.Handle(user, w, r)
		}
		return
	})
}

func mapUser(data map[string]interface{}, m config.UserMapping) (user UserInfo) {
	if id, ok := data[m.ID].(string); ok {
		user.ID = id
	}
	if name, ok := data[m.FirstName].(string); ok {
		user.FirstName = name
	}
	if name, ok := data[m.LastName].(string); ok {
		user.LastName = name
	}
	if email, ok := data[m.EmailAddress].(string); ok {
		user.Email = email
	}
	return
}

func decodeUser(data io.Reader, mapping config.UserMapping) (user UserInfo, err error) {
	var um map[string]interface{}
	if err = json.NewDecoder(data).Decode(&um); err != nil {
		return
	}
	user = mapUser(um, mapping)
	return
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
	if sess, err = h.options.store.Get(r, session.UserSessionName); err != nil {
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
