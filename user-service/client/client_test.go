package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/darren-west/app/user-service/client"
	"github.com/darren-west/app/user-service/models"
	"github.com/stretchr/testify/suite"
)

func TestClientSuite(t *testing.T) {
	suite.Run(t, &ClientSuite{})
}

type ClientSuite struct {
	suite.Suite
}

func (cs *ClientSuite) TestCreateUser() {
	expected := models.UserInfo{ID: "12345", FirstName: "foo", LastName: "bar", Email: "email@email.com"}
	fn := func(r *http.Request) (resp *http.Response, err error) {
		resp = new(http.Response)
		resp.Body = ioutil.NopCloser(&bytes.Buffer{})
		user := cs.decodeUser(r)
		cs.Assert().Equal(expected, user)
		resp.StatusCode = http.StatusCreated
		return
	}
	s := client.New(client.WithRoundTripper(RoundTripFunc(fn)))
	err := s.CreateUser(context.TODO(), expected)
	cs.Assert().NoError(err)
}

func (cs *ClientSuite) TestCreateUserError() {
	fn := func(r *http.Request) (resp *http.Response, err error) {
		resp = new(http.Response)
		cs.Assert().NotZero(r.ContentLength)
		resp.StatusCode = http.StatusInternalServerError
		resp.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("boom")))
		return
	}
	s := client.New(client.WithRoundTripper(RoundTripFunc(fn)))
	err := s.CreateUser(context.TODO(), models.UserInfo{ID: "12345", FirstName: "foo", LastName: "bar", Email: "email@email.com"})
	cs.Assert().EqualError(err, "create user failed: boom, code 500")
}

func (cs *ClientSuite) TestListUsers() {
	expected := []models.UserInfo{
		{
			ID:        "1234",
			FirstName: "foo",
			LastName:  "bar",
			Email:     "email@email.com",
		},
		{
			ID:        "12345",
			FirstName: "foo",
			LastName:  "bar",
			Email:     "email@email.com",
		},
	}
	fn := func(r *http.Request) (resp *http.Response, err error) {
		resp = new(http.Response)
		cs.Assert().Equal(int64(0), r.ContentLength)
		resp.StatusCode = http.StatusOK
		buf := &bytes.Buffer{}
		cs.Require().NoError(json.NewEncoder(buf).Encode(expected))
		resp.Body = ioutil.NopCloser(buf)
		resp.Header = make(map[string][]string)
		resp.Header.Set("Content-Type", "application/json")
		return
	}
	s := client.New(client.WithRoundTripper(RoundTripFunc(fn)))
	users, err := s.ListUsers(context.TODO())
	cs.Assert().NoError(err)
	cs.Assert().Len(users, 2)
	cs.Assert().Equal(expected, users)
}

func (cs ClientSuite) TestUpdateUser() {
	expected := models.UserInfo{ID: "1", FirstName: "foo", LastName: "bar", Email: "email@email.com"}
	fn := func(r *http.Request) (resp *http.Response, err error) {
		cs.Assert().Equal("/users/1", r.URL.Path)
		resp = new(http.Response)
		user := cs.decodeUser(r)
		cs.Assert().Equal(expected, user)
		resp.Body = ioutil.NopCloser(&bytes.Buffer{})
		resp.StatusCode = http.StatusOK
		return
	}
	s := client.New(client.WithRoundTripper(RoundTripFunc(fn)))
	cs.Assert().NoError(s.UpdateUser(context.TODO(), expected))
}

func (cs ClientSuite) TestUpdateUserError() {
	expected := models.UserInfo{ID: "1", FirstName: "foo", LastName: "bar", Email: "email@email.com"}
	fn := func(r *http.Request) (resp *http.Response, err error) {
		resp = new(http.Response)
		resp.StatusCode = http.StatusInternalServerError
		resp.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("boom")))
		return
	}
	s := client.New(client.WithRoundTripper(RoundTripFunc(fn)))
	cs.Assert().EqualError(s.UpdateUser(context.TODO(), expected), "update user failed: boom, code 500")
}

func (cs ClientSuite) TestDeleteUser() {
	expected := "123"
	fn := func(r *http.Request) (resp *http.Response, err error) {
		resp = new(http.Response)
		resp.StatusCode = http.StatusOK
		resp.Body = ioutil.NopCloser(&bytes.Buffer{})
		cs.Assert().Equal("/users/123", r.URL.Path)
		return
	}
	s := client.New(client.WithRoundTripper(RoundTripFunc(fn)))
	cs.Assert().NoError(s.DeleteUser(context.TODO(), expected))
}

func (cs ClientSuite) TestDeleteUserError() {
	expected := "123"
	fn := func(r *http.Request) (resp *http.Response, err error) {
		resp = new(http.Response)
		resp.StatusCode = http.StatusNotFound
		resp.Body = ioutil.NopCloser(bytes.NewBuffer([]byte("fire, fire")))
		return
	}
	s := client.New(client.WithRoundTripper(RoundTripFunc(fn)))
	cs.Assert().EqualError(s.DeleteUser(context.TODO(), expected), "delete user failed: fire, fire, code 404")
}

func (cs ClientSuite) TestGetUser() {
	expected := models.UserInfo{ID: "123", FirstName: "foo", LastName: "bar", Email: "email@email.com"}
	fn := func(r *http.Request) (resp *http.Response, err error) {
		resp = new(http.Response)
		resp.StatusCode = http.StatusOK
		buf := &bytes.Buffer{}
		cs.Require().NoError(json.NewEncoder(buf).Encode(expected))
		resp.Header = make(map[string][]string)
		resp.Header.Set("Content-Type", "application/json")
		resp.Body = ioutil.NopCloser(buf)
		return
	}
	s := client.New(client.WithRoundTripper(RoundTripFunc(fn)))
	user, err := s.GetUser(context.TODO(), expected.ID)
	cs.Assert().NoError(err)

	cs.Assert().Equal(expected, user)
}
func (cs ClientSuite) decodeUser(r *http.Request) (user models.UserInfo) {
	cs.Require().NoError(json.NewDecoder(r.Body).Decode(&user))
	return
}

type RoundTripFunc func(req *http.Request) (*http.Response, error)

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}
