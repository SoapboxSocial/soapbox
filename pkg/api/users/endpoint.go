package users

import (
	"database/sql"
	"log"
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
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	user, err := u.ub.FindByID(i)
	if err != nil {
		if err == sql.ErrNoRows {
			httputil.JsonError(w, 404, httputil.ErrorCodeUserNotFound, "user not found")
			return
		}

		// @todo more specific error
		httputil.JsonError(w, 500, httputil.ErrorCodeFailedToGetUser, "")
		return
	}

	user.Email = nil

	err = httputil.JsonEncode(w, user)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}
