package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/darren-west/app/oauth-service/auth"
	"github.com/darren-west/app/oauth-service/authenticator"
	_ "github.com/darren-west/app/oauth-service/authenticator/google"
	"github.com/darren-west/app/oauth-service/config"
	"github.com/sirupsen/logrus"
)

var (
	configFlag        = flag.String("config", "config.json", "--config the path to the oauth2 configuration file")
	authenticatorFlag = flag.String("authenticator", "google.com", "--authenticator the implementation to use")
)

func init() {
	flag.Parse()
}

func main() {
	authenticator, ok := authenticator.Map[*authenticatorFlag]
	if !ok {
		logrus.Fatalf("implementation %s not imported", *authenticatorFlag)
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
	mux := http.NewServeMux()
	mux.Handle("/", h)
	mux.HandleFunc("/health", health)
	log.Fatal(http.ListenAndServe(":80", mux))
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK) // TODO: think of any unhealthy situations other than the http server not handling traffic.
}
