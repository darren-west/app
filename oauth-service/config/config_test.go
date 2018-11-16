package config_test

import (
	"errors"
	"testing"

	"github.com/darren-west/app/oauth-service/config"
	"github.com/darren-west/app/oauth-service/config/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReader(t *testing.T) {
	testData := `{
		"BindAddress":":80",
		"LoginRoutePath":"/Login",
		"RedirectRoutePath":"/Redirect",
		"APIEndpoint":"www.foo.com",
		"MongoSession": {
			"Name": "session",
			"EncryptKey":"KEY",
			"ConnectionString":"mongodb://database",
			"DatabaseName":"db"
		},
		"OAuth":{
			"ClientID":"foobar", 
			"ClientSecret":"foo", 
			"RedirectURL":"http://redirect.co.uk",
			"Scopes":["scope","scope1"],
			"Endpoint": { 
				"AuthURL":"https://www.facebook.com/dialog/oauth",
				"TokenURL":"https://graph.facebook.com/oauth/access_token"
			}
		},
		"userMapping": {
			"ID": "sub",
			"FirstName": "given_name",
			"LastName": "family_name",
			"EmailAddress": "email"
		}
	}`
	mockFileReader := mocks.NewMockFileReader(gomock.NewController(t))
	mockFileReader.EXPECT().Read("/foo/path.config").Return([]byte(testData), nil)
	reader, err := config.NewReader(mockFileReader)
	require.NoError(t, err)

	conf, err := reader.Read("/foo/path.config")
	require.NoError(t, err)

	assert.Equal(t, ":80", conf.BindAddress)
	assert.Equal(t, "/Login", conf.LoginRoutePath)
	assert.Equal(t, "/Redirect", conf.RedirectRoutePath)

	assert.Equal(t, "KEY", conf.MongoSession.EncryptKey)
	assert.Equal(t, "mongodb://database", conf.MongoSession.ConnectionString)
	assert.Equal(t, "db", conf.MongoSession.DatabaseName)

	assert.Equal(t, "foobar", conf.OAuth.ClientID)
	assert.Equal(t, "foo", conf.OAuth.ClientSecret)
	assert.Equal(t, "http://redirect.co.uk", conf.OAuth.RedirectURL)
	assert.Equal(t, []string{"scope", "scope1"}, conf.OAuth.Scopes)
	assert.Equal(t, "https://www.facebook.com/dialog/oauth", conf.OAuth.Endpoint.AuthURL)
	assert.Equal(t, "https://graph.facebook.com/oauth/access_token", conf.OAuth.Endpoint.TokenURL)
}

func TestReaderCaseInsensitive(t *testing.T) {
	testData := `{
		"bindAddress":":80",
		"loginRoutePath":"/Login",
		"redirectRoutePath":"/Redirect",
		"apiEndpoint":"www.foo.com",
		"mongoSession": {
			"name":"session",
			"encryptKey":"KEY",
			"connectionString":"mongodb://database",
			"databaseName":"db"
		},
		"oAuth":{
			"clientID":"foobar", 
			"clientSecret":"foo", 
			"redirectURL":"http://redirect.co.uk",
			"scopes":["scope","scope1"],
			"endpoint": { 
				"authURL":"https://www.facebook.com/dialog/oauth",
				"tokenURL":"https://graph.facebook.com/oauth/access_token"
			}
		},
		"userMapping": {
			"ID": "sub",
			"FirstName": "given_name",
			"LastName": "family_name",
			"EmailAddress": "email"
		}
	}`
	mockFileReader := mocks.NewMockFileReader(gomock.NewController(t))
	mockFileReader.EXPECT().Read("/foo/path.config").Return([]byte(testData), nil)
	reader, err := config.NewReader(mockFileReader)
	require.NoError(t, err)

	conf, err := reader.Read("/foo/path.config")
	require.NoError(t, err)

	assert.Equal(t, ":80", conf.BindAddress)
	assert.Equal(t, "/Login", conf.LoginRoutePath)
	assert.Equal(t, "/Redirect", conf.RedirectRoutePath)

	assert.Equal(t, "KEY", conf.MongoSession.EncryptKey)
	assert.Equal(t, "mongodb://database", conf.MongoSession.ConnectionString)
	assert.Equal(t, "db", conf.MongoSession.DatabaseName)

	assert.Equal(t, "foobar", conf.OAuth.ClientID)
	assert.Equal(t, "foo", conf.OAuth.ClientSecret)
	assert.Equal(t, "http://redirect.co.uk", conf.OAuth.RedirectURL)
	assert.Equal(t, []string{"scope", "scope1"}, conf.OAuth.Scopes)
	assert.Equal(t, "https://www.facebook.com/dialog/oauth", conf.OAuth.Endpoint.AuthURL)
	assert.Equal(t, "https://graph.facebook.com/oauth/access_token", conf.OAuth.Endpoint.TokenURL)
}

func TestFileReaderError(t *testing.T) {
	mockFileReader := mocks.NewMockFileReader(gomock.NewController(t))
	mockFileReader.EXPECT().Read("/foo/path.config").Return(nil, errors.New("boom"))
	reader, err := config.NewReader(mockFileReader)
	require.NoError(t, err)

	_, err = reader.Read("/foo/path.config")
	assert.EqualError(t, err, "failed to read file: boom")
}

func TestFileReaderNil(t *testing.T) {
	_, err := config.NewReader(nil)
	assert.EqualError(t, err, "file reader is nil")
}

func TestConfigValidation(t *testing.T) {
	testData := `{
		"loginRoutePath":"/Login",
		"redirectRoutePath":"/Redirect",
		"mongoSession": {
			"encryptKey":"KEY",
			"connectionString":"mongodb://database",
			"databaseName":"db"
		},
		"oAuth":{
			"clientID":"foobar", 
			"clientSecret":"foo", 
			"redirectURL":"http://redirect.co.uk",
			"scopes":["scope","scope1"],
			"endpoint": { 
				"authURL":"https://www.facebook.com/dialog/oauth",
				"tokenURL":"https://graph.facebook.com/oauth/access_token"
			}
		},
		"userMapping": {
			"ID": "sub",
			"FirstName": "given_name",
			"LastName": "family_name",
			"EmailAddress": "email"
		}
	}`
	mockFileReader := mocks.NewMockFileReader(gomock.NewController(t))
	mockFileReader.EXPECT().Read("/foo/path.config").Return([]byte(testData), nil)
	reader, err := config.NewReader(mockFileReader)
	require.NoError(t, err)

	_, err = reader.Read("/foo/path.config")
	assert.EqualError(t, err, "configuration invalid: required field bind address missing")
}

func TestConfigValidationMissingOAuth(t *testing.T) {
	testData := `{
		"bindAddress":":80",
		"loginRoutePath":"/Login",
		"redirectRoutePath":"/Redirect",
		"mongoSession": {
			"name": "session",
			"encryptKey":"KEY",
			"connectionString":"mongodb://database",
			"databaseName":"db"
		},
		"userMapping": {
			"ID": "sub",
			"FirstName": "given_name",
			"LastName": "family_name",
			"EmailAddress": "email"
		}
	}`
	mockFileReader := mocks.NewMockFileReader(gomock.NewController(t))
	mockFileReader.EXPECT().Read("/foo/path.config").Return([]byte(testData), nil)
	reader, err := config.NewReader(mockFileReader)
	require.NoError(t, err)

	_, err = reader.Read("/foo/path.config")
	assert.EqualError(t, err, "configuration invalid: required field oauth missing")
}
