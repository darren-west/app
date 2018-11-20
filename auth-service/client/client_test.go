package client_test

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/darren-west/app/auth-service/client"
	"github.com/darren-west/app/utils/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientExchangeTokenSuccess(t *testing.T) {
	fn := RoundTripperFunc(func(req *http.Request) (resp *http.Response, err error) {
		resp = new(http.Response)
		data, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Equal(t, "{\"id\":\"1234\",\"first_name\":\"foo\",\"last_name\":\"bar\",\"email\":\"foo@email.com\"}", string(data))
		resp.StatusCode = http.StatusOK
		resp.Body = ioutil.NopCloser(bytes.NewBufferString("foo"))
		return
	})

	service := client.New(
		client.WithBaseAddress("http://localhost/"),
		client.WithRoundTripper(fn),
	)
	tok, err := service.ExchangeToken(context.Background(), jwt.User{
		ID:        "1234",
		FirstName: "foo",
		LastName:  "bar",
		Email:     "foo@email.com",
	})
	require.NoError(t, err)

	assert.Equal(t, "foo", tok)
}

func TestClientExchangeTokenRetry(t *testing.T) {
	count := 0
	fn := RoundTripperFunc(func(req *http.Request) (resp *http.Response, err error) {
		resp = new(http.Response)
		resp.Body = ioutil.NopCloser(bytes.NewBufferString("foo"))
		if count != 2 {
			count++
			resp.StatusCode = http.StatusInternalServerError
			return
		}
		resp.StatusCode = http.StatusOK
		return
	})

	service := client.New(
		client.WithBaseAddress("http://localhost/"),
		client.WithRoundTripper(fn),
		client.WithRetryCount(3),
	)
	_, err := service.ExchangeToken(context.Background(), jwt.User{
		ID:        "1234",
		FirstName: "foo",
		LastName:  "bar",
		Email:     "foo@email.com",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestClientExchangeTokenError(t *testing.T) {
	fn := RoundTripperFunc(func(req *http.Request) (resp *http.Response, err error) {
		err = errors.New("boom")
		return
	})

	service := client.New(
		client.WithBaseAddress("http://localhost/"),
		client.WithRoundTripper(fn),
	)
	_, err := service.ExchangeToken(context.Background(), jwt.User{
		ID:        "1234",
		FirstName: "foo",
		LastName:  "bar",
		Email:     "foo@email.com",
	})
	require.EqualError(t, err, "exchange token failed: Post http://localhost/token: boom")
}

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
