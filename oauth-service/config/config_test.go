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
		"ClientID":"foobar", 
		"ClientSecret":"foo", 
		"RedirectURL":"http://redirect.co.uk",
		"Scopes":["scope","scope1"],
		"Endpoint": { 
			"AuthURL":"https://www.facebook.com/dialog/oauth",
			"TokenURL":"https://graph.facebook.com/oauth/access_token"
		}
	}`
	mockFileReader := mocks.NewMockFileReader(gomock.NewController(t))
	mockFileReader.EXPECT().Read("/foo/path.config").Return([]byte(testData), nil)
	reader, err := config.NewReader(mockFileReader)
	require.NoError(t, err)

	conf, err := reader.Read("/foo/path.config")
	require.NoError(t, err)

	assert.Equal(t, "foobar", conf.ClientID)
	assert.Equal(t, "foo", conf.ClientSecret)
	assert.Equal(t, "http://redirect.co.uk", conf.RedirectURL)
	assert.Equal(t, []string{"scope", "scope1"}, conf.Scopes)
	assert.Equal(t, "https://www.facebook.com/dialog/oauth", conf.Endpoint.AuthURL)
	assert.Equal(t, "https://graph.facebook.com/oauth/access_token", conf.Endpoint.TokenURL)
}

func TestReaderCaseInsensitive(t *testing.T) {
	testData := `{
		"clientId":"foobar", 
		"clientSecret":"foo", 
		"redirectURL":"http://redirect.co.uk",
		"scopes":["scope","scope1"],
		"endpoint": { 
			"authURL":"https://www.facebook.com/dialog/oauth",
			"tokenURL":"https://graph.facebook.com/oauth/access_token"
		}
	}`
	mockFileReader := mocks.NewMockFileReader(gomock.NewController(t))
	mockFileReader.EXPECT().Read("/foo/path.config").Return([]byte(testData), nil)
	reader, err := config.NewReader(mockFileReader)
	require.NoError(t, err)

	conf, err := reader.Read("/foo/path.config")
	require.NoError(t, err)

	assert.Equal(t, "foobar", conf.ClientID)
	assert.Equal(t, "foo", conf.ClientSecret)
	assert.Equal(t, "http://redirect.co.uk", conf.RedirectURL)
	assert.Equal(t, []string{"scope", "scope1"}, conf.Scopes)
	assert.Equal(t, "https://www.facebook.com/dialog/oauth", conf.Endpoint.AuthURL)
	assert.Equal(t, "https://graph.facebook.com/oauth/access_token", conf.Endpoint.TokenURL)
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
