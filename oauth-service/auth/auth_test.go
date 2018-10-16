package auth_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/darren-west/app/oauth-service/auth"
	"github.com/darren-west/app/oauth-service/auth/mocks"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
)

func TestWithLoginPattern_InvalidOption(t *testing.T) {
	_, err := auth.NewHandler(auth.WithLoginPattern(""))
	assert.EqualError(t, err, "invalid option: login pattern is empty")
}

func TestWithRedirectPattern_InvalidOption(t *testing.T) {
	_, err := auth.NewHandler(auth.WithRedirectPattern(""))
	assert.EqualError(t, err, "invalid option: redirect pattern is empty")
}

func TestPatternOptions_CorrectlySet(t *testing.T) {
	h, err := auth.NewHandler(auth.WithLoginPattern("/foo"), auth.WithRedirectPattern("/bar"))
	require.NoError(t, err)
	assert.Equal(t, "/foo", h.Options().LoginPattern)
	assert.Equal(t, "/bar", h.Options().RedirectPattern)
}

func TestLogin_InvalidStoreOption(t *testing.T) {
	_, err := auth.NewHandler(auth.WithSessionStore(nil))
	assert.EqualError(t, err, "invalid option: store is nil")
}

func TestLogin_InvalidOauthConfig(t *testing.T) {
	_, err := auth.NewHandler(auth.WithOauth2Config(nil))
	assert.EqualError(t, err, "invalid option: oauth config is nil")
}

func TestLogin_InvalidOnAuthenticated(t *testing.T) {
	_, err := auth.NewHandler(auth.WithAuthenticator(nil))
	assert.EqualError(t, err, "invalid option: authenticator is nil")
}

func TestLoginSuite(t *testing.T) {
	suite.Run(t, &LoginSuite{})
}

type LoginSuite struct {
	suite.Suite
	handler auth.Handler

	mockStore *mocks.MockStore

	OAuthConfig *oauth2.Config

	server *httptest.Server
}

func (ls *LoginSuite) SetupTest() {
	var err error
	controller := gomock.NewController(ls.T())
	ls.mockStore = mocks.NewMockStore(controller)

	ls.server = ls.setupEndpoint()

	ls.OAuthConfig = &oauth2.Config{
		ClientID:     "some client id",
		ClientSecret: "some secret id",
		RedirectURL:  "http://127.0.0.1:8080/auth",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/o/oauth2/auth", ls.server.URL),
			TokenURL: fmt.Sprintf("%s/o/oauth2/token", ls.server.URL),
		},
	}
	user := auth.UserInfo{}
	authenticator := mocks.NewMockAuthenticator(controller)
	authenticator.EXPECT().RetrieveUser(gomock.Any()).Return(user, nil)
	authenticator.EXPECT().OnAuthenticated(gomock.Any(), user).Do(func(w http.ResponseWriter, _ auth.UserInfo) {
		fmt.Fprint(w, "authenticated")
	}).Return()

	ls.handler, err = auth.NewHandler(
		auth.WithSessionStore(ls.mockStore),
		auth.WithOauth2Config(ls.OAuthConfig),
		auth.WithAuthenticator(authenticator),
	)
	ls.Require().NoError(err)

}

func (ls *LoginSuite) TeardownTest() {
	ls.server.Close()
}

func (ls *LoginSuite) TestLogin_Success() {
	recorder, request := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/login", nil)
	session := sessions.NewSession(ls.mockStore, ls.handler.Options().SessionName)
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().SessionName).Return(session, nil)
	ls.mockStore.EXPECT().Save(request, recorder, session).Return(nil)

	ls.handler.ServeHTTP(recorder, request)

	ls.Assert().Len(session.Values, 1)
	ls.Assert().NotEmpty(session.Values["state"])

	redirected := recorder.Result().Header.Get("Location") // this is where redirected http request urls are put.
	url, err := url.Parse(redirected)
	ls.Require().NoError(err)
	ls.Assert().Equal("http://127.0.0.1:8080/auth", url.Query().Get("redirect_uri"))
	ls.Assert().Equal("some client id", url.Query().Get("client_id"))
}

func (ls *LoginSuite) TestLogin_SessionError() {
	recorder, request := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/login", nil)
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().SessionName).Return(nil, errors.New("boom"))

	ls.handler.ServeHTTP(recorder, request)

	ls.Assert().Equal("unable to read session: boom\n", recorder.Body.String())
	ls.Assert().Equal(http.StatusInternalServerError, recorder.Code)
}

func (ls *LoginSuite) TestLogin_SessionSaveError() {
	recorder, request := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/login", nil)
	session := sessions.NewSession(ls.mockStore, ls.handler.Options().SessionName)
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().SessionName).Return(session, nil)
	ls.mockStore.EXPECT().Save(request, recorder, session).Return(errors.New("boom"))

	ls.handler.ServeHTTP(recorder, request)

	ls.Assert().Equal("unable to save session: boom\n", recorder.Body.String())
	ls.Assert().Equal(http.StatusInternalServerError, recorder.Code)
}

func (ls *LoginSuite) TestLogin_RedirectInvalidState() {
	recorder, request := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/redirect?state=bar", nil)

	session := sessions.NewSession(ls.mockStore, ls.handler.Options().SessionName)
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().SessionName).Return(session, nil)
	session.Values["state"] = "foo"

	ls.handler.ServeHTTP(recorder, request)

	ls.Assert().Equal("state token invalid\n", recorder.Body.String())
	ls.Assert().Equal(http.StatusUnauthorized, recorder.Code)
}

func (ls *LoginSuite) TestLogin_RedirectSuccess() {
	recorder, request := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/redirect?state=foo&code=blah", nil)

	session := sessions.NewSession(ls.mockStore, ls.handler.Options().SessionName)
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().SessionName).Return(session, nil)
	session.Values["state"] = "foo"

	ls.handler.ServeHTTP(recorder, request)

	ls.Assert().Equal(http.StatusOK, recorder.Code)
	ls.Assert().Equal("authenticated", recorder.Body.String())
}

func (ls *LoginSuite) TestLogin_RedirectOnAuthenticatedCalled() {
	recorder, request := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/redirect?state=foo&code=blah", nil)
	callbackInvoked := false

	authenticator := mocks.NewMockAuthenticator(gomock.NewController(ls.T()))
	user := auth.UserInfo{ID: "foo"}
	authenticator.EXPECT().RetrieveUser(gomock.Any()).Return(user, nil)
	authenticator.EXPECT().OnAuthenticated(recorder, user).Do(func(w http.ResponseWriter, u auth.UserInfo) {
		ls.Assert().NotNil(w)
		ls.Equal(user, u)
		callbackInvoked = true
	}).Return()

	var err error
	ls.handler, err = auth.NewHandler(auth.WithSessionStore(ls.mockStore), auth.WithOauth2Config(ls.OAuthConfig),
		auth.WithAuthenticator(authenticator),
	)
	ls.Require().NoError(err)
	session := sessions.NewSession(ls.mockStore, ls.handler.Options().SessionName)
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().SessionName).Return(session, nil)
	session.Values["state"] = "foo"

	ls.handler.ServeHTTP(recorder, request)

	ls.Assert().Equal(http.StatusOK, recorder.Code)
	ls.Assert().True(callbackInvoked, "the callback has not been invoked")
}

func (ls *LoginSuite) TestLogin_RedirectExchangeError() {
	recorder, request := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/redirect?state=foo&code=blah", nil)

	session := sessions.NewSession(ls.mockStore, ls.handler.Options().SessionName)
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().SessionName).Return(session, nil)
	session.Values["state"] = "foo"
	ls.server.Close()
	ls.handler.ServeHTTP(recorder, request)

	ls.Assert().Equal(http.StatusInternalServerError, recorder.Code)
}

func (ls LoginSuite) setupEndpoint() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/o/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "access_token=%s", "foo")
	})
	return httptest.NewServer(mux)
}
