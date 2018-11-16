package controller_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darren-west/app/user-service/controller"
	"github.com/darren-west/app/user-service/controller/mocks"
	"github.com/darren-west/app/user-service/models"
	"github.com/darren-west/app/user-service/repository"
	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/suite"
)

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, &HandlerSuite{})
}

type HandlerSuite struct {
	suite.Suite
	MockUserRepository *mocks.MockUserRepository
	http.Handler
	*httprouter.Router
}

func (hs *HandlerSuite) SetupTest() {
	hs.MockUserRepository = mocks.NewMockUserRepository(gomock.NewController(hs.T()))
	hs.Handler = controller.NewHandler(hs.MockUserRepository, httprouter.New())
}

func (hs *HandlerSuite) TestGetUser() {
	hs.MockUserRepository.EXPECT().FindUser(repository.NewMatcher().WithID("1234")).Return(models.UserInfo{ID: "1234", FirstName: "foo", LastName: "bar", Email: "foo@email.com"}, nil)

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodGet, "/users/1234", nil)
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusOK, recoder.Code)
	hs.Assert().Equal("{\"ID\":\"1234\",\"FirstName\":\"foo\",\"LastName\":\"bar\",\"Email\":\"foo@email.com\"}\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestGetUserPretty() {
	hs.MockUserRepository.EXPECT().FindUser(repository.NewMatcher().WithID("1234")).Return(models.UserInfo{ID: "1234", FirstName: "foo", LastName: "bar", Email: "foo@email.com"}, nil)

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodGet, "/users/1234?pretty", nil)
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusOK, recoder.Code)
	hs.Assert().Equal("{\n\t\"ID\": \"1234\",\n\t\"FirstName\": \"foo\",\n\t\"LastName\": \"bar\",\n\t\"Email\": \"foo@email.com\"\n}\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestGetUserNotFound() {
	hs.MockUserRepository.EXPECT().FindUser(repository.NewMatcher().WithID("1234")).Return(models.UserInfo{}, errors.New("user not found"))

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodGet, "/users/1234", nil)
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusNotFound, recoder.Code)
	hs.Assert().Equal("user not found\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestGetUserError() {
	hs.MockUserRepository.EXPECT().FindUser(repository.NewMatcher().WithID("1234")).Return(models.UserInfo{}, errors.New("boom"))

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodGet, "/users/1234", nil)
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusInternalServerError, recoder.Code)
	hs.Assert().Equal("boom\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestListUsers() {
	testUsers := []models.UserInfo{
		models.UserInfo{ID: "123", FirstName: "foo", LastName: "bar", Email: "foo@email.com"},
		models.UserInfo{ID: "1234", FirstName: "bar", LastName: "foo", Email: "bar@email.com"},
	}
	hs.MockUserRepository.EXPECT().ListUsers(repository.EmptyMatcher).Return(testUsers, nil)

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodGet, "/users", nil)
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusOK, recoder.Code)
	hs.Assert().Equal("[{\"ID\":\"123\",\"FirstName\":\"foo\",\"LastName\":\"bar\",\"Email\":\"foo@email.com\"},{\"ID\":\"1234\",\"FirstName\":\"bar\",\"LastName\":\"foo\",\"Email\":\"bar@email.com\"}]\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestListUsersPretty() {
	testUsers := []models.UserInfo{
		models.UserInfo{ID: "123", FirstName: "foo", LastName: "bar", Email: "foo@email.com"},
		models.UserInfo{ID: "1234", FirstName: "bar", LastName: "foo", Email: "bar@email.com"},
	}
	hs.MockUserRepository.EXPECT().ListUsers(repository.EmptyMatcher).Return(testUsers, nil)

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodGet, "/users?pretty", nil)
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusOK, recoder.Code)
	hs.Assert().Equal("[\n\t{\n\t\t\"ID\": \"123\",\n\t\t\"FirstName\": \"foo\",\n\t\t\"LastName\": \"bar\",\n\t\t\"Email\": \"foo@email.com\"\n\t},\n\t{\n\t\t\"ID\": \"1234\",\n\t\t\"FirstName\": \"bar\",\n\t\t\"LastName\": \"foo\",\n\t\t\"Email\": \"bar@email.com\"\n\t}\n]\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestListUsersError() {
	hs.MockUserRepository.EXPECT().ListUsers(repository.EmptyMatcher).Return(nil, errors.New("boom"))

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodGet, "/users", nil)
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusInternalServerError, recoder.Code)
	hs.Assert().Equal("boom\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestListUsersEmpty() {
	hs.MockUserRepository.EXPECT().ListUsers(repository.EmptyMatcher).Return([]models.UserInfo{}, nil)

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodGet, "/users", nil)
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusOK, recoder.Code)
	hs.Assert().Equal("[]\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestDeleteUser() {
	hs.MockUserRepository.EXPECT().RemoveUser(repository.NewMatcher().WithID("12345")).Return(nil)

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodDelete, "/users/12345", nil)
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusOK, recoder.Code)
	hs.Assert().Equal("", recoder.Body.String())
}

func (hs *HandlerSuite) TestDeleteUserError() {
	hs.MockUserRepository.EXPECT().RemoveUser(repository.NewMatcher().WithID("12345")).Return(errors.New("boom"))

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodDelete, "/users/12345", nil)
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusInternalServerError, recoder.Code)
	hs.Assert().Equal("boom\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestUpdateUser() {
	user := models.UserInfo{ID: "12345", FirstName: "foo", LastName: "bar", Email: "email@email.com"}
	hs.MockUserRepository.EXPECT().UpdateUser(user).Return(nil)

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodPut, "/users/12345", hs.EncodeUser(user))
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusOK, recoder.Code)
	hs.Assert().Equal("", recoder.Body.String())
}

func (hs *HandlerSuite) TestUpdateUserError() {
	user := models.UserInfo{ID: "12345", FirstName: "foo", LastName: "bar", Email: "email@email.com"}
	hs.MockUserRepository.EXPECT().UpdateUser(user).Return(errors.New("big explosion"))

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodPut, "/users/12345", hs.EncodeUser(user))
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusInternalServerError, recoder.Code)
	hs.Assert().Equal("big explosion\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestUpdateUserInvalid() {
	user := models.UserInfo{ID: "12345", FirstName: "foo", LastName: "bar", Email: ""}

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodPut, "/users/12345", hs.EncodeUser(user))
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusBadRequest, recoder.Code)
	hs.Assert().Equal("invalid user, missing email\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestUpdateUserNotFound() {
	user := models.UserInfo{ID: "12345", FirstName: "foo", LastName: "bar", Email: "email@email.com"}
	hs.MockUserRepository.EXPECT().UpdateUser(user).Return(errors.New("user not found"))

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodPut, "/users/12345", hs.EncodeUser(user))
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusNotFound, recoder.Code)
	hs.Assert().Equal("user not found\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestUpdateUserNotMatchingID() {
	user := models.UserInfo{ID: "12345", FirstName: "foo", LastName: "bar", Email: "email@email.com"}

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodPut, "/users/123458", hs.EncodeUser(user))
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusBadRequest, recoder.Code)
	hs.Assert().Equal("query param id (12345) does not match request body id (123458)\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestCreateUser() {
	user := models.UserInfo{ID: "12345", FirstName: "foo", LastName: "bar", Email: "email@email.com"}
	hs.MockUserRepository.EXPECT().CreateUser(user).Return(nil)

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodPost, "/users", hs.EncodeUser(user))
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusCreated, recoder.Code)
	hs.Assert().Equal("", recoder.Body.String())
}

func (hs *HandlerSuite) TestCreateUserError() {
	user := models.UserInfo{ID: "12345", FirstName: "foo", LastName: "bar", Email: "email@email.com"}
	hs.MockUserRepository.EXPECT().CreateUser(user).Return(errors.New("look away"))

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodPost, "/users", hs.EncodeUser(user))
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusInternalServerError, recoder.Code)
	hs.Assert().Equal("look away\n", recoder.Body.String())
}

func (hs *HandlerSuite) TestCreateUserInvalid() {
	user := models.UserInfo{ID: "", FirstName: "foo", LastName: "bar", Email: "email@email.com"}

	recoder := httptest.NewRecorder()
	request := hs.NewRequest(http.MethodPost, "/users", hs.EncodeUser(user))
	hs.Handler.ServeHTTP(recoder, request)

	hs.Assert().Equal(http.StatusBadRequest, recoder.Code)
	hs.Assert().Equal("invalid user, missing id\n", recoder.Body.String())
}

func (hs *HandlerSuite) NewRequest(method string, path string, r io.Reader) *http.Request {
	req := httptest.NewRequest(method, path, r)
	req.Header.Set("Content-Type", "application/json")
	return req
}

func (hs *HandlerSuite) EncodeUser(user models.UserInfo) (buf *bytes.Buffer) {
	buf = new(bytes.Buffer)
	hs.Require().NoError(json.NewEncoder(buf).Encode(&user))
	return
}
