package devices

import (
	"log"
	"net/http"

	auth "github.com/ephemeral-networks/voicely/pkg/api/middleware"
	"github.com/ephemeral-networks/voicely/pkg/devices"
	httputil "github.com/ephemeral-networks/voicely/pkg/http"
)

type DevicesEndpoint struct {
	db *devices.DevicesBackend
}

func NewDevicesEndpoint(db *devices.DevicesBackend) *DevicesEndpoint {
	return &DevicesEndpoint{
		db: db,
	}
}

func (d *DevicesEndpoint) AddDevice(w http.ResponseWriter, r *http.Request) {
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
	if err != nil {
		log.Println("failed to create session: ", err.Error())
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToStoreDevice, "failed")
		return
	}

	httputil.JsonSuccess(w)
}
