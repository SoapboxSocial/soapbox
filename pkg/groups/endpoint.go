package groups

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	auth "github.com/soapboxsocial/soapbox/pkg/api/middleware"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/images"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
)

type Endpoint struct {
	backend *Backend
	images  *images.Backend
	queue   *pubsub.Queue
}

func NewEndpoint(backend *Backend, ib *images.Backend, queue *pubsub.Queue) *Endpoint {
	return &Endpoint{
		backend: backend,
		images:  ib,
		queue:   queue,
	}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.Path("/create").Methods("POST").HandlerFunc(e.CreateGroup)
	r.Path("/{id:[0-9]+}").Methods("GET").HandlerFunc(e.GetGroup)
	r.Path("/{id:[0-9]+}/edit").Methods("POST").HandlerFunc(e.EditGroup)
	r.Path("/{id:[0-9]+}/invite").Methods("GET").HandlerFunc(e.GetUserInviteForGroup)
	r.Path("/{id:[0-9]+}/invite").Methods("POST").HandlerFunc(e.InviteUsersToGroup)
	r.Path("/{id:[0-9]+}/members").Methods("GET").HandlerFunc(e.GetGroupMembers)

	// @TODO NOT SURE IF I AM HAPPY WITH THESE
	r.Path("/{id:[0-9]+}/invite/decline").Methods("POST").HandlerFunc(e.DeclineInvite)
	r.Path("/{id:[0-9]+}/invite/accept").Methods("POST").HandlerFunc(e.AcceptInvite)

	r.Path("/{id:[0-9]+}/join").Methods("POST").HandlerFunc(e.JoinGroup)

	return r
}

func (e *Endpoint) CreateGroup(w http.ResponseWriter, r *http.Request) {
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

	name := strings.TrimSpace(r.Form.Get("name"))
	if name == "" || len(name) > 256 {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid name")
		return
	}

	description := strings.TrimSpace(strings.ReplaceAll(r.Form.Get("description"), "\n", " "))
	if len([]rune(description)) > 300 {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "description too long")
		return
	}

	groupType := strings.TrimSpace(r.Form.Get("group_type"))
	if groupType != "public" && groupType != "private" && groupType != "restricted" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid group")
		return
	}

	img, err := e.handleGroupImage(r)
	if err != nil && err != http.ErrMissingFile {
		fmt.Println(err)
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	id, err := e.backend.CreateGroup(userID, name, description, img, groupType)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "failed to create")
		return
	}

	_ = e.queue.Publish(pubsub.GroupTopic, pubsub.NewGroupCreationEvent(id, userID, name))

	err = httputil.JsonEncode(w, map[string]interface{}{"success": true, "id": id})
	if err != nil {
		log.Println("error writing response: " + err.Error())
	}
}

func (e *Endpoint) GetGroupsForUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	limit := httputil.GetInt(r.URL.Query(), "limit", 10)
	offset := httputil.GetInt(r.URL.Query(), "offset", 0)

	var result []*Group

	if userID == id {
		result, err = e.backend.GetGroupsForUser(id, limit, offset)
	} else {
		result, err = e.backend.GetGroupsForProfile(id, limit, offset)
	}

	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetFollowers, "")
		return
	}

	err = httputil.JsonEncode(w, result)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

func (e *Endpoint) GetGroup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	result, err := e.backend.GetGroupForUser(userID, id)
	if err != nil {
		if err == sql.ErrNoRows {
			httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeNotFound, "not found")
			return
		}

		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetFollowers, "")
		return
	}

	err = httputil.JsonEncode(w, result)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

func (e *Endpoint) InviteUsersToGroup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	idsStr := r.Form.Get("ids")
	ids := strings.Split(idsStr, ",")
	if len(ids) == 0 {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "missing ids")
		return
	}

	group, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid group")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	ok, err = e.backend.IsAdminForGroup(userID, group)
	if !ok {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeUnauthorized, "unauthorized")
		return
	}

	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "server error")
		return
	}

	for _, id := range ids {
		go func(val string) {
			id, err := strconv.Atoi(val)
			if err != nil {
				return
			}

			err = e.backend.InviteUser(userID, group, id)
			if err != nil {
				return
			}

			_ = e.queue.Publish(pubsub.GroupTopic, pubsub.NewGroupInviteEvent(userID, id, group))
		}(id)
	}

	httputil.JsonSuccess(w)
}

func (e *Endpoint) GetUserInviteForGroup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	group, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid group")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	result, err := e.backend.GetInviterForUser(userID, group)
	if err != nil {
		if err == sql.ErrNoRows {
			httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeNotFound, "not found")
			return
		}

		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetFollowers, "")
		return
	}

	err = httputil.JsonEncode(w, result)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

func (e *Endpoint) DeclineInvite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	group, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid group")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = e.backend.DeclineInvite(userID, group)
	if err != nil {
		// @TODO BETTER
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	httputil.JsonSuccess(w)
}

func (e *Endpoint) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	group, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid group")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = e.backend.AcceptInvite(userID, group)
	if err != nil {
		// @TODO BETTER
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	httputil.JsonSuccess(w)
}

func (e *Endpoint) JoinGroup(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	group, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid group")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	public, err := e.backend.IsPublic(group)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	if !public {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = e.backend.Join(userID, group)
	if err != nil {
		// @TODO BETTER
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	httputil.JsonSuccess(w)
}

func (e *Endpoint) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	limit := httputil.GetInt(r.URL.Query(), "limit", 10)
	offset := httputil.GetInt(r.URL.Query(), "offset", 0)

	result, err := e.backend.GetAllMembers(id, limit, offset)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetFollowers, "")
		return
	}

	err = httputil.JsonEncode(w, result)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

func (e *Endpoint) EditGroup(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	params := mux.Vars(r)
	group, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	isAdmin, err := e.backend.IsAdminForGroup(userID, group)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	if !isAdmin {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeUnauthorized, "invalid id")
		return
	}

	description := strings.TrimSpace(strings.ReplaceAll(r.Form.Get("description"), "\n", " "))
	if len([]rune(description)) > 300 {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	oldPath, err := e.backend.GetGroupImage(group)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	image := oldPath
	path, err := e.handleGroupImage(r)
	if err != nil && err != http.ErrMissingFile {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	if path != "" {
		image = path
	}

	err = e.backend.UpdateGroup(group, description, image)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	if image != oldPath {
		_ = e.images.Remove(oldPath)
	}

	err = e.queue.Publish(pubsub.GroupTopic, pubsub.NewGroupUpdateEvent(group))
	if err != nil {
		log.Printf("queue.Publish err: %v\n", err)
	}

	httputil.JsonSuccess(w)
}

func (e *Endpoint) handleGroupImage(r *http.Request) (string, error) {
	file, _, err := r.FormFile("image")
	if err != nil {
		return "", err
	}

	image, err := images.MultipartFileToPng(file)
	if err != nil {
		return "", err
	}

	name, err := e.images.StoreGroupPhoto(image)
	if err != nil {
		return "", err
	}

	return name, nil
}
