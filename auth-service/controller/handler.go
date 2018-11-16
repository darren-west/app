package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/darren-west/app/utils/jwt"

	"github.com/darren-west/app/utils/httputil"

	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	jwt.Writer
}

func NewHandler(keyPath string, router *httprouter.Router) http.Handler {
	h := Handler{
		Writer: jwt.NewWriter(jwt.WriterBuilder.WithPrivateKeyPath(keyPath)),
	}
	router.POST("/token", httputil.UseErrorHandle(h.ExchangeToken))
	return router
}

func (h Handler) ExchangeToken(w http.ResponseWriter, r *http.Request, _ httprouter.Params) httputil.Error {
	user := jwt.User{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return httputil.NewError(http.StatusInternalServerError).WithError(err)
	}
	token, err := h.Writer.Write(&jwt.Claims{
		User:      user,
		ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		IssuedAt:  time.Now().Unix(),
	})
	if err != nil {
		return httputil.NewError(http.StatusInternalServerError).WithError(err)
	}
	fmt.Fprintf(w, "%s", token)
	return nil
}
