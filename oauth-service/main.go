package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/darren-west/app/utils/fileutil"
	"github.com/darren-west/app/utils/session"

	"github.com/darren-west/app/oauth-service/auth"
	"github.com/darren-west/app/oauth-service/config"
	"github.com/darren-west/app/oauth-service/redirector"
	"github.com/sirupsen/logrus"
)

var (
	configFlag = flag.String("config", "config.json", "--config the path to the oauth2 configuration file")
)

func init() {
	flag.Parse()
}

func main() {
	reader, err := config.NewReader(fileutil.FileReader{})
	if err != nil {
		logrus.Fatal(err)
	}
	config, err := reader.Read(*configFlag)
	if err != nil {
		logrus.Fatal(err)
	}

	store, err := session.NewMongoStore(config.MongoSession)
	if err != nil {
		logrus.Fatal(err)
	}

	h, err := auth.NewHandler(
		auth.WithConfig(config),
		auth.WithSessionStore(store),
		auth.WithLoginHandler(redirector.Login{Store: store}),
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
