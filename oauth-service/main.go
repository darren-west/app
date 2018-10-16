package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/darren-west/app/oauth-service/auth"
	"github.com/darren-west/app/oauth-service/config"

	_ "github.com/darren-west/app/oauth-service/authenticator/google"

	"github.com/darren-west/app/oauth-service/authenticator"
	"github.com/sirupsen/logrus"
)

var (
	configFlag        = flag.String("config", "config.json", "--config the path to the oauth2 configuration file")
	authenticatorFlag = flag.String("authenticator", "google", "--authenticator the implementation to use")
)

func init() {
	flag.Parse()
}

func main() {
	authenticator, ok := authenticator.Map[*authenticatorFlag]
	if !ok {
		logrus.Fatalf("implementation %s not imported", authenticatorFlag)
	}
	reader, err := config.NewReader(config.DefaultFileReader{})
	if err != nil {
		logrus.Fatal(err)
	}
	config, err := reader.Read(*configFlag)
	if err != nil {
		logrus.Fatal(err)
	}

	h, err := auth.NewHandler(
		auth.WithOauth2Config(config),
		auth.WithAuthenticator(authenticator),
		auth.WithRedirectPattern("/auth"),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", h))

}
