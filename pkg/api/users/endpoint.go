package users

import (
	"database/sql"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	auth "github.com/soapboxsocial/soapbox/pkg/api/middleware"
	"github.com/soapboxsocial/soapbox/pkg/followers"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/images"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type UsersEndpoint struct {
	ub          *users.UserBackend
	fb          *followers.FollowersBackend
	sm          *sessions.SessionManager
	ib          *images.Backend
	currentRoom *rooms.CurrentRoomBackend

	queue *pubsub.Queue
}

func NewUsersEndpoint(
	ub *users.UserBackend,
	fb *followers.FollowersBackend,
	sm *sessions.SessionManager,
	ib *images.Backend,
	queue *pubsub.Queue,
	cr *rooms.CurrentRoomBackend,
) *UsersEndpoint {
	return &UsersEndpoint{
		ub:          ub,
		fb:          fb,
		sm:          sm,
		ib:          ib,
		queue:       queue,
		currentRoom: cr,
	}
}

func (u *UsersEndpoint) GetUserByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	u.handleUserRetrieval(id, w, r)
}

func (u *UsersEndpoint) GetUsersByUsername(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	username := params["username"]

	id, err := u.ub.GetIDForUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeUserNotFound, "user not found")
			return
		}

		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	u.handleUserRetrieval(id, w, r)
}

func (u *UsersEndpoint) handleUserRetrieval(id int, w http.ResponseWriter, r *http.Request) {
	caller, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	var user *users.Profile
	var err error

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

	cr, err := u.currentRoom.GetCurrentRoomForUser(id)
	if err != nil && err.Error() != "redis: nil" {
		log.Println("current room retrieval error", err.Error())
	}

	if cr != 0 {
		user.CurrentRoom = &cr
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

	limit := httputil.GetInt(r.URL.Query(), "limit", 10)
	offset := httputil.GetInt(r.URL.Query(), "offset", 0)

	result, err := u.fb.GetAllUsersFollowing(id, limit, offset)
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

	limit := httputil.GetInt(r.URL.Query(), "limit", 10)
	offset := httputil.GetInt(r.URL.Query(), "offset", 0)

	result, err := u.fb.GetAllUsersFollowedBy(id, limit, offset)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetFollowers, "")
		return
	}

	err = httputil.JsonEncode(w, result)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

func (u *UsersEndpoint) GetMyFriends(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	result, err := u.fb.GetFriends(id)
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

	err = u.follow(userID, id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "failed to follow")
		return
	}

	httputil.JsonSuccess(w)
}

// @TODO, ERROR HANDLER DOESN'T SEEM NICE HERE
func (u *UsersEndpoint) MultiFollowUsers(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	ids := strings.Split(r.Form.Get("ids"), ",")
	for _, raw := range ids {
		id, err := strconv.Atoi(raw)
		if err != nil {
			httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
			return
		}

		err = u.follow(userID, id)
		if err != nil {
			continue
		}
	}

	httputil.JsonSuccess(w)
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
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	name := strings.TrimSpace(r.Form.Get("display_name"))
	if name == "" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	bio := strings.TrimSpace(strings.ReplaceAll(r.Form.Get("bio"), "\n", " "))
	if len([]rune(bio)) > 300 {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	oldPath, err := u.ub.GetProfileImage(userID)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	file, _, err := r.FormFile("profile")
	if err != nil && err != http.ErrMissingFile {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	image := oldPath
	if file != nil {
		image, err = u.processProfilePicture(file)
		if err != nil {
			httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
			return
		}
	}

	err = u.ub.UpdateUser(userID, name, bio, image)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	if image != oldPath {
		_ = u.ib.Remove(oldPath)
	}

	err = u.queue.Publish(pubsub.UserTopic, pubsub.NewUserUpdateEvent(userID))
	if err != nil {
		log.Printf("queue.Publish err: %v\n", err)
	}

	httputil.JsonSuccess(w)
}

func (u *UsersEndpoint) follow(userID, id int) error {
	err := u.fb.FollowUser(userID, id)
	if err != nil {
		return err
	}

	go func() {
		err = u.queue.Publish(pubsub.UserTopic, pubsub.NewFollowerEvent(userID, id))
		if err != nil {
			log.Printf("queue.Publish err: %v\n", err)
		}
	}()

	return nil
}

func (u *UsersEndpoint) processProfilePicture(file multipart.File) (string, error) {
	pngBytes, err := images.MultipartFileToPng(file)
	if err != nil {
		return "", err
	}

	name, err := u.ib.Store(pngBytes)
	if err != nil {
		return "", err
	}

	return name, nil
}
