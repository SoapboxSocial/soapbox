package metadata

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Endpoint struct {
	usersBackend users.UserBackend
}

func NewEndpoint(usersBackend users.UserBackend) *Endpoint  {
	return &Endpoint{
		usersBackend: usersBackend,
	}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/users/{username}", e.user).Methods("GET")

	return r
}

func (e *Endpoint) user(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	username := params["username"]
	user, err := e.usersBackend.GetUserByUsername(username)
	if err != nil {
		httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeNotFound, "not found")
		return
	}

	err = httputil.JsonEncode(w, user)
	if err != nil {
		log.Printf("failed to encode: %v", err)
	}
}
