package main

import (
	"net/http"

	"github.com/darren-west/app/auth-service/controller"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

func main() {
	router := httprouter.New()
	if err := http.ListenAndServe(":80", controller.NewHandler("../utils/jwt/testdata/app.rsa", router)); err != nil {
		logrus.Error(err)
	}
}
