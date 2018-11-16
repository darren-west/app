package client_test

import (
	"context"
	"testing"

	"github.com/darren-west/app/auth-service/client"
	"github.com/darren-west/app/user-service/models"
)

func TestClient(t *testing.T) {
	s := client.New("http://devel.com/api/auth")

	tok, err := s.ExchangeToken(context.Background(), models.UserInfo{
		ID:        "123",
		FirstName: "Foo",
		LastName:  "bar",
		Email:     "email@email.com",
	})
	t.Log(err)

	t.Log(tok)
}
