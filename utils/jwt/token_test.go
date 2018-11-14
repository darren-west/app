package jwt_test

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/darren-west/app/utils/fileutil"
	"github.com/darren-west/app/utils/jwt"
	"github.com/darren-west/app/utils/jwt/mocks"
)

func TestTokenWrite(t *testing.T) {
	w := jwt.NewWriter(
		jwt.WriterBuilder.WithFileReader(fileutil.FileReader{}),
		jwt.WriterBuilder.WithPrivateKeyPath("testdata/app.rsa"),
	)

	token, err := w.Write(&jwt.Claims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
		User:      jwt.User{ID: "1234", FirstName: "foo", LastName: "bar", Email: "foo@email.com"},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	data := strings.Split(token.String(), ".")
	require.Len(t, data, 3)
	decodedString, err := base64.StdEncoding.DecodeString(data[1])
	require.NoError(t, err)
	payload := struct {
		User struct {
			ID        string `json:"id"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
		} `json:"user"`
	}{}
	require.NoError(t, json.Unmarshal(decodedString, &payload))

	assert.Equal(t, "1234", payload.User.ID)
	assert.Equal(t, "foo", payload.User.FirstName)
	assert.Equal(t, "bar", payload.User.LastName)
	assert.Equal(t, "foo@email.com", payload.User.Email)
}

func TestWriterOptions(t *testing.T) {
	w := jwt.NewWriter(jwt.WriterBuilder.WithPrivateKeyPath("/foo"))
	assert.Equal(t, "/foo", w.PrivateKeyPath())

	mock := mocks.NewMockFileReader(gomock.NewController(t))
	w = jwt.NewWriter(jwt.WriterBuilder.WithFileReader(mock))

	assert.Equal(t, mock, w.FileReader())
}

func TestWriterReaderNil(t *testing.T) {
	defer func() {
		r := recover()
		assert.EqualError(t, r.(error), "file reader is nil")
	}()
	jwt.NewWriter(jwt.WriterBuilder.WithFileReader(nil))
}

func TestReader(t *testing.T) {
	r := jwt.NewReader(
		jwt.ReaderBuilder.WithPublicKeyPath("testdata/app.rsa.pub"),
		jwt.ReaderBuilder.WithFileReader(fileutil.FileReader{}),
	)

	claims, err := r.Read(jwt.NewToken("eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJ1c2VyIjp7ImlkIjoiMTIzNCIsImZpcnN0X25hbWUiOiJmb28iLCJsYXN0X25hbWUiOiJiYXIiLCJlbWFpbCI6ImZvb0BlbWFpbC5jb20ifSwiZXhwIjoxNTQxOTMyMjA1LCJpYXQiOjE1NDE4NDU4MDV9.ZY3jOt5NiiKEoKm7fJh8gd_OMk1pvisF31Xt0rX5X3xYj8mDkXhGRjp7xMnKtvyviINeKXJ1N2x6JwvCkBOprC6K1GMULpMA3DVXLpGp5PHRHQg7Yck0i03I-2rCHK6m2yJEkAQkmqhb44g34bzurlv7049dveioYKtvPEvtA_xOUWawVU4jKv3fhPu5ZK0edHw6x_PLXNExI9BfPFPYY_-8ZjWIVmBIbZCeQInswAbQQsJWMDJbp-_6TNIRkDzfZwls7wd70fsnxBMfWzGvI3gk323CkLhUhV0FxahYAvrjq3LbgNLLSE9TgPbaIwQ5XCO2o6RViojWwqLBWgNzwA"))
	assert.NoError(t, err)

	assert.Equal(t, "foo", claims.User.FirstName)
	assert.Equal(t, "bar", claims.User.LastName)
	assert.Equal(t, "foo@email.com", claims.User.Email)
	assert.NotZero(t, claims.IssuedAt)
	assert.NotZero(t, claims.ExpiresAt)
}

func TestReaderOptions(t *testing.T) {
	r := jwt.NewReader(jwt.ReaderBuilder.WithPublicKeyPath("/foo"))
	assert.Equal(t, "/foo", r.PublicKeyPath())

	mock := mocks.NewMockFileReader(gomock.NewController(t))
	r = jwt.NewReader(jwt.ReaderBuilder.WithFileReader(mock))

	assert.Equal(t, mock, r.FileReader())
}
