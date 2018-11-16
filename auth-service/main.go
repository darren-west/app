package main

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

func main() {
	handler := negroni.New()
	router := httprouter.New()
	//router.POST("/exchange/token/:id", AuthenticationHandler{}.ExchangeToken)
	handler.UseHandler(router)
	if err := http.ListenAndServe(":80", handler); err != nil {
		logrus.Error(err)
	}
}
