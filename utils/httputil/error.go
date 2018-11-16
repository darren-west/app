package httputil

import (
	"fmt"
	"net/http"

	"github.com/darren-west/app/utils/httputil"
	"github.com/julienschmidt/httprouter"
)

type (
	Error interface {
		error
		StatusCode() int
		Write(w http.ResponseWriter)
	}

	ErrorBuilder interface {
		Error
		WithMessage(format string, a ...interface{}) Error
		WithError(err error) Error
	}

	httpErr struct {
		message    string
		statusCode int
	}
)

func NewError(statusCode int) ErrorBuilder {
	return &httpErr{
		statusCode: statusCode,
	}
}

func (e *httpErr) Write(w http.ResponseWriter) {
	http.Error(w, e.Error(), e.StatusCode())
}

func (e *httpErr) Error() string {
	return e.message
}

func (e *httpErr) StatusCode() int {
	return e.statusCode
}

func (e *httpErr) WithMessage(format string, a ...interface{}) Error {
	e.message = fmt.Sprintf(format, a...)
	return e
}

func (e *httpErr) WithError(err error) Error {
	return e.WithMessage(err.Error())
}

type ErrorHandle func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (err httputil.Error)

func UseErrorHandle(f ErrorHandle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if err := f(w, r, ps); err != nil {
			err.Write(w)
			return
		}

	}
}
