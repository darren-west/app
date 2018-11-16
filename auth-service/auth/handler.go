package auth

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Handler struct {
}

func (Handler) ExchangeToken(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	user := User{}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id := ps.ByName("id")
	if id != user.ID {
		http.Error(w, "param id does not match body id", http.StatusInternalServerError)
		return
	}

}

type User struct {
	ID        string
	FirstName string
	LastName  string
	Email     string
}

// user exists
// oauth-service -> user -> auth -> token

// user doesnt exist
// oauth-service -> user -> create -> auth -> token
