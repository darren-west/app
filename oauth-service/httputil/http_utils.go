package httputil

import (
	"fmt"
	"net/http"
)

// Error a simple structure to simplify http errors.
type Error struct {
	code    int
	message string
}

// StatusCode returns the http status code of the error.
func (e *Error) StatusCode() int {
	return e.code
}

// Error returns a formated error message.
func (e *Error) Error() string {
	return fmt.Sprintf("http status %d: %s", e.code, e.message)
}

// Write writes the error to a http response writer.
func (e *Error) Write(w http.ResponseWriter) {
	http.Error(w, e.message, e.code)
}

// NewError returns a new error setting the http status code and internal error.
func NewError(statusCode int, err error) *Error {
	message := ""
	if err != nil {
		message = err.Error()
	}
	return &Error{
		code:    statusCode,
		message: message,
	}
}
