package stories

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
)

type Endpoint struct {
	backend *Backend
}

func NewEndpoint(backend *Backend) *Endpoint {
	return &Endpoint{backend: backend}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.Path("/upload").Methods("POST").HandlerFunc(e.UploadStory)
	r.Path("/delete/{id:[0-9]+}").Methods("DELETE").HandlerFunc(e.DeleteStory)

	return r
}

func (e *Endpoint) UploadStory(w http.ResponseWriter, r *http.Request) {
	//userID, ok := auth.GetUserIDFromContext(r.Context())
	//if !ok {
	//	httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
	//	return
	//}
}

func (e *Endpoint) DeleteStory(w http.ResponseWriter, r *http.Request) {

}

func (e *Endpoint) GetStoriesForUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	stories, err := e.backend.GetStoriesForUser(id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = httputil.JsonEncode(w, stories)
	if err != nil {
		log.Printf("failed to write story response: %s\n", err.Error())
	}
}
