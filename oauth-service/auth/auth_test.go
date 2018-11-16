package auth_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/darren-west/app/oauth-service/auth"
	"github.com/darren-west/app/oauth-service/auth/mocks"
	"github.com/darren-west/app/oauth-service/config"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
)

func TestLogin_InvalidStoreOption(t *testing.T) {
	_, err := auth.NewHandler(auth.WithSessionStore(nil))
	assert.EqualError(t, err, "invalid option: store is nil")
}

func TestLoginSuite(t *testing.T) {
	suite.Run(t, &LoginSuite{})
}

type LoginSuite struct {
	suite.Suite
	handler auth.Handler

	mockStore *mocks.MockStore

	mockLoginHandler *mocks.MockLoginHandler

	OAuthConfig *oauth2.Config

	server *httptest.Server

	config.Options
}

func (ls *LoginSuite) SetupTest() {
	var err error
	controller := gomock.NewController(ls.T())
	ls.mockStore = mocks.NewMockStore(controller)
	ls.mockLoginHandler = mocks.NewMockLoginHandler(controller)

	ls.server = ls.setupEndpoint()

	ls.Options = config.Options{
		OAuth: &oauth2.Config{
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
		},
		LoginRoutePath:    "/login",
		RedirectRoutePath: "/redirect",
		BindAddress:       ":80",
	}

	ls.handler, err = auth.NewHandler(
		auth.WithSessionStore(ls.mockStore),
		auth.WithConfig(ls.Options),
		auth.WithLoginHandler(ls.mockLoginHandler),
	)
	ls.Require().NoError(err)

}

func (ls *LoginSuite) TeardownTest() {
	ls.server.Close()
}

func (ls *LoginSuite) TestLogin_Success() {
	recorder, request := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/login", nil)
	session := sessions.NewSession(ls.mockStore, ls.handler.Options().Config.MongoSession.Name)
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().Config.MongoSession.Name).Return(session, nil)
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
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().Config.MongoSession.Name).Return(nil, errors.New("boom"))

	ls.handler.ServeHTTP(recorder, request)

	ls.Assert().Equal("unable to read session: boom\n", recorder.Body.String())
	ls.Assert().Equal(http.StatusInternalServerError, recorder.Code)
}

func (ls *LoginSuite) TestLogin_SessionSaveError() {
	recorder, request := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/login", nil)
	session := sessions.NewSession(ls.mockStore, ls.handler.Options().Config.MongoSession.Name)
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().Config.MongoSession.Name).Return(session, nil)
	ls.mockStore.EXPECT().Save(request, recorder, session).Return(errors.New("boom"))

	ls.handler.ServeHTTP(recorder, request)

	ls.Assert().Equal("unable to save session: boom\n", recorder.Body.String())
	ls.Assert().Equal(http.StatusInternalServerError, recorder.Code)
}

func (ls *LoginSuite) TestLogin_RedirectInvalidState() {
	recorder, request := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/redirect?state=bar", nil)

	session := sessions.NewSession(ls.mockStore, ls.handler.Options().Config.MongoSession.Name)
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().Config.MongoSession.Name).Return(session, nil)
	session.Values["state"] = "foo"

	ls.handler.ServeHTTP(recorder, request)

	ls.Assert().Equal("state token invalid\n", recorder.Body.String())
	ls.Assert().Equal(http.StatusUnauthorized, recorder.Code)
}

func (ls *LoginSuite) TestLogin_RedirectSuccess() {
	recorder, request := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/redirect?state=foo&code=blah", nil)

	session := sessions.NewSession(ls.mockStore, ls.handler.Options().Config.MongoSession.Name)
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().Config.MongoSession.Name).Return(session, nil)
	ls.mockLoginHandler.EXPECT().Handle(gomock.Any(), recorder, request).Return()
	session.Values["state"] = "foo"

	ls.handler.ServeHTTP(recorder, request)

	ls.Assert().Equal(http.StatusOK, recorder.Code)

}

func (ls *LoginSuite) TestLogin_RedirectExchangeError() {
	recorder, request := httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/redirect?state=foo&code=blah", nil)

	session := sessions.NewSession(ls.mockStore, ls.handler.Options().Config.MongoSession.Name)
	ls.mockStore.EXPECT().Get(request, ls.handler.Options().Config.MongoSession.Name).Return(session, nil)
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
