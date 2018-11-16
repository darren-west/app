package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darren-west/app/auth-service/auth"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/suite"
)

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, &HandlerSuite{})
}

type HandlerSuite struct {
	suite.Suite

	authHandler auth.Handler
}

func (s *HandlerSuite) SetupTest() {

}

func (s *HandlerSuite) TestAuthenticationToken() {
	recorder := httptest.NewRecorder()

	user := auth.OAuthUser{
		ID:        "1234",
		FirstName: "foo",
		LastName:  "bar",
		Email:     "foo@email.com",
	}

	buf := &bytes.Buffer{}
	s.Assert().NoError(json.NewEncoder(buf).Encode(&user))

	request := httptest.NewRequest(http.MethodPost, "/exchange/token/1234", buf)

	s.authHandler.ExchangeToken(recorder, request, []httprouter.Param{httprouter.Param{Key: "id", Value: "1234"}})

	s.Assert().Equal(http.StatusOK, recorder.Code)

	token := struct{ Token string }{}

	s.Assert().NoError(json.NewDecoder(recorder.Body).Decode(&token))
}
