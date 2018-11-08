package session

import (
	"fmt"
	"time"

	"github.com/kidstuff/mongostore"
	mgo "gopkg.in/mgo.v2"
)

const (
	UserSessionName = "user-data"
)

func NewMongoStore(options Options) (store *mongostore.MongoStore, err error) {
	session, err := mgo.Dial(options.ConnectionString)
	if err != nil {
		return
	}
	store = mongostore.NewMongoStore(
		session.DB(options.DatabaseName).C("SessionData"),
		int(options.MaxAge),
		options.EnsureTTL,
		[]byte(options.EncryptKey),
	)
	return
}

type Options struct {
	ConnectionString string
	DatabaseName     string
	MaxAge           time.Duration
	EnsureTTL        bool
	EncryptKey       string
}

func (o Options) IsValid() (err error) {
	merr := "mongo session invalid: %s"
	if o.EncryptKey == "" {
		err = fmt.Errorf(merr, "encryption key cannot be empty")
		return
	}
	if o.ConnectionString == "" {
		err = fmt.Errorf(merr, "connection string cannot be empty")
		return
	}
	if o.DatabaseName == "" {
		err = fmt.Errorf(merr, "database name cannot be empty")
		return
	}
	return
}
