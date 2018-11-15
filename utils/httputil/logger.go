package httputil

import (
	"net/http"
	"net/http/httputil"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// WithHandlerLogging adds a log message for every request/response handled.
func WithHandlerLogging(l *logrus.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		decorateLogWithRequest(l, r).WithField("id", id).Debug("incoming request")
		h.ServeHTTP(w, r)
		l.WithField("id", id).Debug("outgoing response")
	})
}

func decorateLogWithRequest(l *logrus.Logger, r *http.Request) *logrus.Entry {
	data, _ := httputil.DumpRequest(r, true)
	return l.WithField("http request", string(data))
}
