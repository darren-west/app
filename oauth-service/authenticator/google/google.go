package google

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/darren-west/app/oauth-service/authenticator"

	"github.com/darren-west/app/oauth-service/auth"
)

func init() {
	authenticator.Map["google.com"] = Authenticator{}
}

// Authenticator is used to retrieve google user information using an oauth2 authenticated client.
type Authenticator struct{}

func (Authenticator) RetrieveUser(client *http.Client) (user auth.UserInfo, err error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return
	}
	rawUser := struct {
		ID        string `json:"sub"`
		FirstName string `json:"given_name"`
		LastName  string `json:"family_name"`
		Email     string `json:"email"`
	}{}
	if err = json.NewDecoder(resp.Body).Decode(&rawUser); err != nil {
		return
	}
	user.ID = rawUser.ID
	user.FirstName = rawUser.FirstName
	user.LastName = rawUser.LastName
	user.Email = rawUser.Email
	return
}

// OnAuthenticated handles an authenticated user.
func (Authenticator) OnAuthenticated(w http.ResponseWriter, user auth.UserInfo) {
	fmt.Fprintf(w, "%v", user)
}
