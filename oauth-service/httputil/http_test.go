package httputil_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/darren-west/oauth-service/httputil"
	"github.com/stretchr/testify/assert"
)

func TestHttpErrorMessage(t *testing.T) {
	err := httputil.NewError(http.StatusInternalServerError, errors.New("boom"))
	assert.EqualError(t, err, "http status 500: boom")
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode())
}

func TestHttpError(t *testing.T) {
	recorder := httptest.NewRecorder()
	httputil.NewError(http.StatusUnauthorized, errors.New("foo is not authorized")).Write(recorder)
	assert.Equal(t, "foo is not authorized\n", recorder.Body.String())
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestHttpErrorNilError(t *testing.T) {
	recorder := httptest.NewRecorder()
	httputil.NewError(http.StatusUnauthorized, nil).Write(recorder)
	assert.Equal(t, "\n", recorder.Body.String())
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
