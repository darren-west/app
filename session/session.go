package session

import (
	"os"
	"time"

	"github.com/kidstuff/mongostore"
	mgo "gopkg.in/mgo.v2"
)

const CollectionName = "AuthenticationSession"

func NewMongoStore(options Options) (store *mongostore.MongoStore, err error) {
	session, err := mgo.Dial(options.MongoConnectionString)
	if err != nil {
		return
	}
	store = mongostore.NewMongoStore(
		session.DB(options.DatabaseName).C(CollectionName),
		int(options.MaxAge),
		options.EnsureTTL,
		[]byte(os.Getenv(options.KeyEnvName)),
	)
	return
}

type Options struct {
	MongoConnectionString string
	DatabaseName          string
	MaxAge                time.Duration
	EnsureTTL             bool
	KeyEnvName            string
}
