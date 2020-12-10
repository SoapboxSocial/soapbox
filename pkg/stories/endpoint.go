package stories

import (
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	auth "github.com/soapboxsocial/soapbox/pkg/api/middleware"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
)

type Endpoint struct {
	backend *Backend
	files   *FileBackend
}

func NewEndpoint(backend *Backend, files *FileBackend) *Endpoint {
	return &Endpoint{backend: backend, files: files}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.Path("/upload").Methods("POST").HandlerFunc(e.UploadStory)
	r.Path("/{id:[0-9]+}").Methods("DELETE").HandlerFunc(e.DeleteStory)

	return r
}

func (e *Endpoint) UploadStory(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(2 << 20)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	timestamp, err := strconv.ParseInt(r.Form.Get("device_timestamp"), 10, 64)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	expires := time.Now().Add(24 * time.Hour).Unix()

	file, _, err := r.FormFile("story")
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "no story")
		return
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "failed to upload")
		return
	}

	name, err := e.files.Store(bytes)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "no story")
		return
	}

	err = e.backend.AddStory(IDFromName(name), userID, expires, timestamp)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "no story")
		return
	}

	// @TODO CLEANUP
	httputil.JsonSuccess(w)
}

func (e *Endpoint) DeleteStory(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id := params["id"]
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err := e.backend.DeleteStory(id, userID)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	// @TODO trigger event to delete from file.
}

func (e *Endpoint) GetStoriesForUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	stories, err := e.backend.GetStoriesForUser(id, time.Now().Unix())
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = httputil.JsonEncode(w, stories)
	if err != nil {
		log.Printf("failed to write story response: %s\n", err.Error())
	}
}

func IDFromName(name string) string {
	return strings.TrimSuffix(name, ".aac")
}
