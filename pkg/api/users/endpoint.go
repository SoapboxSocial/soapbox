package users

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	httputil "github.com/ephemeral-networks/voicely/pkg/http"
	"github.com/ephemeral-networks/voicely/pkg/users"
)

type UsersEndpoint struct {
	ub *users.UserBackend
}

func NewUsersEndpoint(ub *users.UserBackend) *UsersEndpoint {
	return &UsersEndpoint{ub: ub}
}

func (u *UsersEndpoint) GetUserByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	i, err := strconv.Atoi(id)
	if err != nil {
		// @todo
		return
	}

	user, err := u.ub.FindByID(i)
	if err != nil {
		// @todo
		return
	}

	user.Email = nil

	err = httputil.JsonEncode(w, user)
	if err != nil {
		// @todo log
	}
}
