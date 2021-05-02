package analytics

import (
	"net/http"

	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
)

type Endpoint struct {
	backend *Backend
}

func NewEndpoint(backend *Backend) *Endpoint {
	return &Endpoint{
		backend: backend,
	}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/notifications/{id}/opened", e.openedNotification).Methods("POST")

	return r
}

func (e *Endpoint) openedNotification(w http.ResponseWriter, r *http.Request) {
	userID, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	params := mux.Vars(r)

	id := params["id"]
	if id == "" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid")
		return
	}

	err := e.backend.MarkNotificationRead(userID, id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeMissingParameter, "invalid")
		return
	}

	httputil.JsonSuccess(w)
}