package main

import (
	"net/http"

	"github.com/darren-west/app/auth-service/controller"
	"github.com/sirupsen/logrus"

	"github.com/julienschmidt/httprouter"
)

func main() {
	router := httprouter.New()
	if err := http.ListenAndServe(":80", controller.NewHandler(router)); err != nil {
		logrus.Error(err)
	}
}
