package users

import (
	"database/sql"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/soapboxsocial/soapbox/pkg/followers"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/images"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/stories"
)

type Endpoint struct {
	ub          *Backend
	fb          *followers.FollowersBackend
	sm          *sessions.SessionManager
	ib          *images.Backend
	stories     *stories.Backend

	queue *pubsub.Queue
}

func NewEndpoint(
	ub *Backend,
	fb *followers.FollowersBackend,
	sm *sessions.SessionManager,
	ib *images.Backend,
	queue *pubsub.Queue,
	stories *stories.Backend,
) *Endpoint {
	return &Endpoint{
		ub:          ub,
		fb:          fb,
		sm:          sm,
		ib:          ib,
		queue:       queue,
		stories:     stories,
	}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.Path("/{id:[0-9]+}").Methods("GET").HandlerFunc(e.GetUserByID)
	r.Path("/{username:[a-z0-9_]+}").Methods("GET").HandlerFunc(e.GetUserByUsername)
	r.Path("/{id:[0-9]+}/followers").Methods("GET").HandlerFunc(e.GetFollowersForUser)
	r.Path("/{id:[0-9]+}/following").Methods("GET").HandlerFunc(e.GetFollowedByForUser)
	r.Path("/{id:[0-9]+}/friends").Methods("GET").HandlerFunc(e.GetFriends)
	r.Path("/follow").Methods("POST").HandlerFunc(e.FollowUser)
	r.Path("/unfollow").Methods("POST").HandlerFunc(e.UnfollowUser)
	r.Path("/multi-follow").Methods("POST").HandlerFunc(e.MultiFollowUsers)
	r.Path("/edit").Methods("POST").HandlerFunc(e.EditUser)
	r.Path("/{id:[0-9]+}/stories").Methods("GET").HandlerFunc(e.GetStoriesForUser)

	return r
}

func (e *Endpoint) GetUserByID(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	e.handleUserRetrieval(id, w, r)
}

func (e *Endpoint) GetUserByUsername(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	username := params["username"]

	id, err := e.ub.GetIDForUsername(username)
	if err != nil {
		if err == sql.ErrNoRows {
			httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeUserNotFound, "user not found")
			return
		}

		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	e.handleUserRetrieval(id, w, r)
}

func (e *Endpoint) handleUserRetrieval(id int, w http.ResponseWriter, r *http.Request) {
	caller, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	var user *Profile
	var err error

	if caller == id {
		user, err = e.ub.GetMyProfile(id)
	} else {
		user, err = e.ub.ProfileByID(id, caller)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeUserNotFound, "user not found")
			return
		}

		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetUser, "")
		return
	}

	// @TODO READD LATER
	//cr, err := e.currentRoom.GetCurrentRoomForUser(id)
	//if err != nil && err.Error() != "redis: nil" {
	//	log.Println("current room retrieval error", err.Error())
	//}
	//
	//if cr != "" {
	//	user.CurrentRoom = &cr
	//}

	err = httputil.JsonEncode(w, user)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

// @todo think about moving these 2 endpoints into a follower specific thing?
func (e *Endpoint) GetFollowersForUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	limit := httputil.GetInt(r.URL.Query(), "limit", 10)
	offset := httputil.GetInt(r.URL.Query(), "offset", 0)

	result, err := e.fb.GetAllUsersFollowing(id, limit, offset)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetFollowers, "")
		return
	}

	err = httputil.JsonEncode(w, result)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

func (e *Endpoint) GetFollowedByForUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	limit := httputil.GetInt(r.URL.Query(), "limit", 10)
	offset := httputil.GetInt(r.URL.Query(), "offset", 0)

	result, err := e.fb.GetAllUsersFollowedBy(id, limit, offset)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetFollowers, "")
		return
	}

	err = httputil.JsonEncode(w, result)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

func (e *Endpoint) GetFriends(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	result, err := e.fb.GetFriends(id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetFollowers, "")
		return
	}

	err = httputil.JsonEncode(w, result)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

func (e *Endpoint) FollowUser(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = e.follow(userID, id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "failed to follow")
		return
	}

	httputil.JsonSuccess(w)
}

// @TODO, ERROR HANDLER DOESN'T SEEM NICE HERE
func (e *Endpoint) MultiFollowUsers(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	userID, ok := httputil.GetUserIDFromContext(r.Context())
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

		err = e.follow(userID, id)
		if err != nil {
			continue
		}
	}

	httputil.JsonSuccess(w)
}

func (e *Endpoint) UnfollowUser(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = e.fb.UnfollowUser(userID, id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "failed to unfollow")
		return
	}

	httputil.JsonSuccess(w)
}

func (e *Endpoint) EditUser(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	userID, ok := httputil.GetUserIDFromContext(r.Context())
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
	if len([]rune(bio)) > 150 {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	oldPath, err := e.ub.GetProfileImage(userID)
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
		image, err = e.processProfilePicture(file)
		if err != nil {
			httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
			return
		}
	}

	err = e.ub.UpdateUser(userID, name, bio, image)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	if image != oldPath {
		_ = e.ib.Remove(oldPath)
	}

	err = e.queue.Publish(pubsub.UserTopic, pubsub.NewUserUpdateEvent(userID))
	if err != nil {
		log.Printf("queue.Publish err: %v\n", err)
	}

	httputil.JsonSuccess(w)
}

func (e *Endpoint) GetStoriesForUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	s, err := e.stories.GetStoriesForUser(id, time.Now().Unix())
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = httputil.JsonEncode(w, s)
	if err != nil {
		log.Printf("failed to write story response: %s\n", err.Error())
	}
}

func (e *Endpoint) follow(userID, id int) error {
	err := e.fb.FollowUser(userID, id)
	if err != nil {
		return err
	}

	go func() {
		err = e.queue.Publish(pubsub.UserTopic, pubsub.NewFollowerEvent(userID, id))
		if err != nil {
			log.Printf("queue.Publish err: %v\n", err)
		}
	}()

	return nil
}

func (e *Endpoint) processProfilePicture(file multipart.File) (string, error) {
	pngBytes, err := images.MultipartFileToPng(file)
	if err != nil {
		return "", err
	}

	name, err := e.ib.Store(pngBytes)
	if err != nil {
		return "", err
	}

	return name, nil
}
