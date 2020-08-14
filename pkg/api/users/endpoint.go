package users

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	auth "github.com/ephemeral-networks/voicely/pkg/api/middleware"
	"github.com/ephemeral-networks/voicely/pkg/followers"
	httputil "github.com/ephemeral-networks/voicely/pkg/http"
	"github.com/ephemeral-networks/voicely/pkg/sessions"
	"github.com/ephemeral-networks/voicely/pkg/users"
)

type UsersEndpoint struct {
	ub *users.UserBackend
	fb *followers.FollowersBackend
	sm *sessions.SessionManager
}

func NewUsersEndpoint(ub *users.UserBackend, fb *followers.FollowersBackend, sm *sessions.SessionManager) *UsersEndpoint {
	return &UsersEndpoint{ub: ub, fb: fb, sm: sm}
}

func (u *UsersEndpoint) GetUserByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	caller, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	var user *users.Profile
	if caller == id {
		user, err = u.ub.GetMyProfile(id)
	} else {
		user, err = u.ub.ProfileByID(id, caller)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			httputil.JsonError(w, 404, httputil.ErrorCodeUserNotFound, "user not found")
			return
		}

		httputil.JsonError(w, 500, httputil.ErrorCodeFailedToGetUser, "")
		return
	}

	err = httputil.JsonEncode(w, user)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

// @todo think about moving these 2 endpoints into a follower specific thing?
func (u *UsersEndpoint) GetFollowersForUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	result, err := u.fb.GetAllUsersFollowing(id)
	if err != nil {
		httputil.JsonError(w, 500, httputil.ErrorCodeFailedToGetFollowers, "")
		return
	}

	err = httputil.JsonEncode(w, result)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

func (u *UsersEndpoint) GetFollowedByForUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	result, err := u.fb.GetAllUsersFollowedBy(id)
	if err != nil {
		httputil.JsonError(w, 500, httputil.ErrorCodeFailedToGetFollowers, "")
		return
	}

	err = httputil.JsonEncode(w, result)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

func (u *UsersEndpoint) FollowUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	id, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	userID, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = u.fb.FollowUser(userID, id)
	if err != nil {
		httputil.JsonError(w, 500, httputil.ErrorCodeInvalidRequestBody, "failed to follow")
		return
	}

	httputil.JsonSuccess(w)
}

func (u *UsersEndpoint) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	id, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	userID, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = u.fb.UnfollowUser(userID, id)
	if err != nil {
		httputil.JsonError(w, 500, httputil.ErrorCodeInvalidRequestBody, "failed to unfollow")
		return
	}

	httputil.JsonSuccess(w)
}
