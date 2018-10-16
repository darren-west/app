package facebook

import (
	"fmt"
	"net/http"

	"github.com/darren-west/app/oauth-service/auth"
	"github.com/darren-west/app/oauth-service/authenticator"
)

func init() {
	authenticator.Map["facebook.com"] = Authenticator{}
}

type Authenticator struct{}

func (Authenticator) RetrieveUser(client *http.Client) (user auth.UserInfo, err error) {
	return
}

func (Authenticator) OnAuthenticated(w http.ResponseWriter, user auth.UserInfo) {
	fmt.Fprintf(w, "%v", user)
}
