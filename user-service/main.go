package main

import (
	"net/http"

	"github.com/darren-west/app/utils/httputil"

	"github.com/darren-west/app/user-service/controller"
	"github.com/darren-west/app/user-service/repository"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

func main() {
	repo, err := repository.NewMongoUserRepository(
		repository.WithConnectionString("mongodb://localhost:27017"),
		repository.WithDatabaseName("dev"),
		repository.WithCollectionName("users"),
	)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.SetLevel(logrus.DebugLevel)

	router := httprouter.New()
	http.ListenAndServe(":80",
		httputil.WithHandlerLogging(logrus.StandardLogger(), controller.NewHandler(repo, router)),
	)
}
