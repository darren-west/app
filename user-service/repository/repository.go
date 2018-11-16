package repository

import (
	"errors"
	"time"

	"github.com/darren-west/app/user-service/models"
	"gopkg.in/mgo.v2"

	"github.com/hashicorp/errwrap"
)

var EmptyMatcher = NewMatcher()

type Matcher map[string]interface{}

func (m Matcher) WithID(id string) Matcher {
	m["id"] = id
	return m
}

func (m Matcher) WithFirstName(name string) Matcher {
	m["name"] = name
	return m
}

func NewMatcher() Matcher {
	return Matcher(make(map[string]interface{}))
}

type Options struct {
	ConnectionString string
	DatabaseName     string
	CollectionName   string
}

type MongoUserRepository struct {
	options Options
	session *mgo.Session
}

func IsErrDuplicateUser(err error) bool {
	return mgo.IsDup(err)
}

func IsErrUserNotFound(err error) bool {
	return errwrap.Contains(err, "user not found")
}

func (r MongoUserRepository) FindUser(m Matcher) (user models.UserInfo, err error) {
	err = r.run(func(c *mgo.Collection) error {
		return c.Find(m).One(&user)
	})
	if err == mgo.ErrNotFound {
		err = errwrap.Wrap(errors.New("user not found"), err)
		return
	}
	return
}

func (r MongoUserRepository) ListUsers(m Matcher) (users []models.UserInfo, err error) {
	err = r.run(func(c *mgo.Collection) error {
		return c.Find(m).All(&users)
	})
	return
}

func (r MongoUserRepository) CreateUser(user models.UserInfo) (err error) {
	err = r.run(func(c *mgo.Collection) error {
		return c.Insert(&user)
	})
	return
}

func (r MongoUserRepository) RemoveUser(m Matcher) (err error) {
	err = r.run(func(c *mgo.Collection) error {
		return c.Remove(m)
	})
	if err == mgo.ErrNotFound {
		err = errwrap.Wrap(errors.New("user not found"), err)
		return
	}
	return
}

func (r MongoUserRepository) UpdateUser(user models.UserInfo) (err error) {
	m := NewMatcher().WithID(user.ID)
	err = r.run(func(c *mgo.Collection) error {
		return c.Update(m, &user)
	})
	if err == mgo.ErrNotFound {
		err = errwrap.Wrap(errors.New("user not found"), err)
		return
	}
	return
}

func (r MongoUserRepository) Options() Options {
	return r.options
}

func (r MongoUserRepository) run(f func(c *mgo.Collection) error) error {
	session := r.session.Clone()
	defer session.Close()
	return f(session.DB(r.options.DatabaseName).C(r.options.CollectionName))
}

func NewMongoUserRepository(opts ...Option) (repo MongoUserRepository, err error) {
	repo = MongoUserRepository{
		options: Options{},
	}
	for _, opt := range opts {
		opt(&repo.options)
	}
	session, err := mgo.DialWithTimeout(repo.options.ConnectionString, time.Second*30)
	if err != nil {
		return
	}
	repo.session = session
	session.SetMode(mgo.Monotonic, true)
	err = session.DB(repo.options.DatabaseName).C(repo.options.CollectionName).EnsureIndex(mgo.Index{
		Key:    []string{"id"},
		Unique: true,
	})

	return
}

type Option func(*Options) error

func WithConnectionString(connectionString string) Option {
	return func(o *Options) (err error) {
		o.ConnectionString = connectionString
		return
	}
}

func WithDatabaseName(db string) Option {
	return func(o *Options) (err error) {
		o.DatabaseName = db
		return
	}
}

func WithCollectionName(c string) Option {
	return func(o *Options) (err error) {
		o.CollectionName = c
		return
	}
}
