package users

import (
	"database/sql"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	auth "github.com/ephemeral-networks/voicely/pkg/api/middleware"
	"github.com/ephemeral-networks/voicely/pkg/followers"
	httputil "github.com/ephemeral-networks/voicely/pkg/http"
	"github.com/ephemeral-networks/voicely/pkg/images"
	"github.com/ephemeral-networks/voicely/pkg/notifications"
	"github.com/ephemeral-networks/voicely/pkg/sessions"
	"github.com/ephemeral-networks/voicely/pkg/users"
)

type UsersEndpoint struct {
	ub *users.UserBackend
	fb *followers.FollowersBackend
	sm *sessions.SessionManager
	ib *images.Backend

	queue *notifications.Queue
}

func NewUsersEndpoint(
	ub *users.UserBackend,
	fb *followers.FollowersBackend,
	sm *sessions.SessionManager,
	queue *notifications.Queue,
	ib *images.Backend,
) *UsersEndpoint {
	return &UsersEndpoint{ub: ub, fb: fb, sm: sm, ib: ib, queue: queue}
}

func (u *UsersEndpoint) GetUserByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	caller, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
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
			httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeUserNotFound, "user not found")
			return
		}

		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetUser, "")
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
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	result, err := u.fb.GetAllUsersFollowing(id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetFollowers, "")
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
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	result, err := u.fb.GetAllUsersFollowedBy(id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetFollowers, "")
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
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	id, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = u.fb.FollowUser(userID, id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "failed to follow")
		return
	}

	httputil.JsonSuccess(w)

	u.queue.Push(notifications.Event{
		Type:    notifications.EventTypeNewFollower,
		Creator: userID,
		Params:  map[string]interface{}{"id": id},
	})
}

func (u *UsersEndpoint) UnfollowUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	id, err := strconv.Atoi(r.Form.Get("id"))
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = u.fb.UnfollowUser(userID, id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "failed to unfollow")
		return
	}

	httputil.JsonSuccess(w)
}

func (u *UsersEndpoint) EditUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "kek")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	name := r.Form.Get("display_name")
	if name == "" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	file, _, err := r.FormFile("profile")
	if err != nil && err != http.ErrMissingFile {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	image := ""
	if file != nil {
		image, err = u.processProfilePicture(file)
		if err != nil {
			httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
			return
		}

		// @todo fine old file name and remove it after update
	}

	err = u.ub.UpdateUser(userID, name, image)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	httputil.JsonSuccess(w)
}

func (u *UsersEndpoint) processProfilePicture(file multipart.File) (string, error) {
	imgBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	pngBytes, err := images.ToPNG(imgBytes)
	if err != nil {
		return "", err
	}

	name, err := u.ib.Store(pngBytes)
	if err != nil {
		return "", err
	}

	return name, nil
}
