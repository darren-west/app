package httputil_test

import (
	"net/http"
	"testing"

	"github.com/darren-west/app/utils/httputil"
	"github.com/stretchr/testify/assert"
)

func TestError(t *testing.T) {
	err := httputil.NewError(http.StatusNotFound).WithMessage("not found")
	assert.Equal(t, err.Error(), "not found")

	assert.Equal(t, http.StatusNotFound, err.StatusCode())
}
