package groups

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	auth "github.com/soapboxsocial/soapbox/pkg/api/middleware"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/images"
)

type Endpoint struct {
	backend *Backend
	images  *images.Backend
}

func NewEndpoint(backend *Backend, ib *images.Backend) *Endpoint {
	return &Endpoint{
		backend: backend,
		images:  ib,
	}
}

// @TODO DO WHAT THIS GUY DID
func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.Path("/create").Methods("POST").HandlerFunc(e.CreateGroup)
	r.Path("/{id:[0-9]+}/invite").Methods("POST").HandlerFunc(e.InviteUsersToGroup)

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
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	id, err := e.backend.CreateGroup(userID, name, description, img, groupType)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "failed to create")
		return
	}

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

	limit := httputil.GetInt(r.URL.Query(), "limit", 10)
	offset := httputil.GetInt(r.URL.Query(), "offset", 0)

	result, err := e.backend.GetGroupsForUser(id, limit, offset)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetFollowers, "")
		return
	}

	err = httputil.JsonEncode(w, result)
	if err != nil {
		log.Printf("failed to write user response: %s\n", err.Error())
	}
}

func (e *Endpoint) InviteUsersToGroup(w http.ResponseWriter, r *http.Request) {
	// @TODO, ENSURE:
	//   - caller is admin
	//   - caller is followed by all users he is inviting.
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
