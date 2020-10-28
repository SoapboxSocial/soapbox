package devices

import (
	"net/http"

	"github.com/gorilla/mux"

	auth "github.com/soapboxsocial/soapbox/pkg/api/middleware"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
)

type Endpoint struct {
	db *Backend
}

func NewEndpoint(db *Backend) *Endpoint {
	return &Endpoint{
		db: db,
	}
}

func (d *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/add", d.add).Methods("POST")

	return r
}

func (d *Endpoint) add(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	token := r.Form.Get("token")
	if token == "" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid token")
		return
	}

	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err = d.db.AddDeviceForUser(userID, token)
	if err != nil && err.Error() != "pq: duplicate key value violates unique constraint \"devices_pkey\"" {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToStoreDevice, "failed")
		return
	}

	httputil.JsonSuccess(w)
}
