// +build integration

package repository_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"gopkg.in/mgo.v2/bson"

	"gopkg.in/mgo.v2"

	"github.com/darren-west/app/user-service/models"
	"github.com/darren-west/app/user-service/repository/testutils/container"

	"github.com/darren-west/app/user-service/repository"
	"github.com/stretchr/testify/suite"
)

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, &RepositorySuite{})
}

type RepositorySuite struct {
	suite.Suite
	repo    repository.MongoUserRepository
	manager *container.Manager
	session *mgo.Session
}

func (rs *RepositorySuite) SetupTest() {
	manager, err := container.NewManager(
		container.WithPortMapping(container.PortMapping{}.With("27017/tcp", "27017/tcp")),
		container.WithImage("mongo:latest"),
		container.WithWriter(ioutil.Discard),
	)
	rs.Require().NoError(err)
	rs.manager = manager
	rs.Require().NoError(rs.manager.Start(context.Background()))
	repo, err := repository.NewMongoUserRepository(
		repository.WithConnectionString("mongodb://127.0.0.1:27017"),
		repository.WithDatabaseName("test"),
		repository.WithCollectionName("users"),
	)
	rs.Require().NoError(err)
	rs.repo = repo
	rs.session, err = mgo.Dial("mongodb://127.0.0.1:27017")
	rs.Require().NoError(err)
}

func (rs *RepositorySuite) TearDownTest() {
	rs.session.Close()
	rs.Require().NoError(rs.manager.Stop(context.Background()))
}

func (rs *RepositorySuite) TestFindUser() {
	expectedUser := models.UserInfo{ID: "1234", FirstName: "foo", LastName: "bar", Email: "foo@email.com"}
	rs.collection().Insert(&expectedUser)

	user, err := rs.repo.FindUser(repository.NewMatcher().WithID("1234"))
	rs.Require().NoError(err)
	rs.Assert().Equal(expectedUser, user)
}

func (rs *RepositorySuite) TestFindUserNotFound() {
	user, err := rs.repo.FindUser(repository.NewMatcher().WithID("1234"))
	rs.Assert().True(repository.IsErrUserNotFound(err))
	rs.Assert().Zero(user)
}

func (rs *RepositorySuite) collection() *mgo.Collection {
	return rs.session.DB(rs.repo.Options().DatabaseName).C(rs.repo.Options().CollectionName)
}

func (rs *RepositorySuite) TestCreateUser() {
	expectedUser := models.UserInfo{ID: "123", FirstName: "foo", LastName: "bar", Email: "foo@email.com"}
	rs.Require().NoError(rs.repo.CreateUser(expectedUser))

	user := models.UserInfo{}
	rs.Require().NoError(rs.collection().Find(bson.M{"id": "123"}).One(&user))
	rs.Assert().Equal(expectedUser, user)
}

func (rs *RepositorySuite) TestRemoveUser() {
	expectedUser := models.UserInfo{ID: "1234", FirstName: "foo", LastName: "bar", Email: "foo@email.com"}
	rs.Require().NoError(rs.collection().Insert(&expectedUser))

	rs.Require().NoError(rs.repo.RemoveUser(repository.NewMatcher().WithID("1234")))

	count, err := rs.collection().Find(bson.M{"id": "1234"}).Count()
	rs.Require().NoError(err)
	rs.Assert().Equal(count, 0)
}

func (rs *RepositorySuite) TestUpdateUser() {
	user := models.UserInfo{ID: "12345", FirstName: "foo", LastName: "bar", Email: "foo@email.com"}
	rs.Require().NoError(rs.collection().Insert(user))

	user.FirstName = "bar"
	user.LastName = "foo"
	user.Email = "foo@email.com"
	rs.Assert().NoError(rs.repo.UpdateUser(user))
	rs.Assert().Equal(models.UserInfo{ID: "12345", FirstName: "bar", LastName: "foo", Email: "foo@email.com"}, user)
}

func (rs *RepositorySuite) TestUpdateUserNotFound() {
	user := models.UserInfo{ID: "12345", FirstName: "foo", LastName: "bar", Email: "foo@email.com"}
	rs.Assert().True(repository.IsErrUserNotFound(rs.repo.UpdateUser(user)))
}

func (rs *RepositorySuite) TestRemoveUserNotFound() {
	err := rs.repo.RemoveUser(repository.NewMatcher().WithID("1234"))
	rs.Assert().True(repository.IsErrUserNotFound(err))
}

func (rs *RepositorySuite) TestListUser() {
	for i := 0; i < 100; i++ {
		rs.Require().NoError(rs.collection().Insert(&models.UserInfo{ID: fmt.Sprintf("%d", i), FirstName: "foo", LastName: "bar", Email: "foo@email.com"}))
	}

	users, err := rs.repo.ListUsers(repository.EmptyMatcher)
	rs.Require().NoError(err)
	rs.Len(users, 100)
}

func (rs *RepositorySuite) TestListUserMatcher() {
	for i := 0; i < 50; i++ {
		rs.Require().NoError(rs.collection().Insert(&models.UserInfo{ID: fmt.Sprintf("%d", i), FirstName: "foo", LastName: "bar", Email: "foo@email.com"}))
	}

	users, err := rs.repo.ListUsers(repository.NewMatcher().WithID("1"))
	rs.Require().NoError(err)
	rs.Len(users, 1)
}

func (rs *RepositorySuite) TestListUsersNone() {
	users, err := rs.repo.ListUsers(repository.EmptyMatcher)
	rs.Assert().NoError(err)
	rs.Assert().Len(users, 0)
}

func (rs *RepositorySuite) TestIndexCreated() {
	indexs, err := rs.collection().Indexes()
	rs.Require().NoError(err)

	rs.Require().Len(indexs, 2)
	rs.Assert().Len(indexs[1].Key, 1)
	rs.Assert().Equal("id", indexs[1].Key[0])
	rs.Assert().Equal(true, indexs[1].Unique)
}
