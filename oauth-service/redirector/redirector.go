package redirector

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/darren-west/app/oauth-service/auth"
	"github.com/darren-west/app/oauth-service/httputil"
	"github.com/darren-west/app/utils/session"
	"github.com/gorilla/sessions"
)

var _ auth.LoginHandler = Login{} // ensure the Login handler implements the interface.

type Login struct {
	Store sessions.Store
}

func (l Login) Handle(user auth.UserInfo, w http.ResponseWriter, r *http.Request) {
	session, err := l.Store.Get(r, session.UserSessionName)
	if err != nil {
		httputil.NewError(http.StatusInternalServerError, err).Write(w)
		return
	}
	delete(session.Values, "state")
	session.Values["api-token"] = base64.StdEncoding.EncodeToString([]byte(user.FirstName))

	data, err := json.Marshal(&user)
	if err != nil {
		httputil.NewError(http.StatusInternalServerError, err).Write(w)
		return
	}

	session.Values["user"] = data

	if err = session.Save(r, w); err != nil {
		httputil.NewError(http.StatusInternalServerError, err).Write(w)
		return
	}

	http.Redirect(w, r, "app/", http.StatusPermanentRedirect)
	// TODO: write username + name to cookie so it can be read by js.

}
