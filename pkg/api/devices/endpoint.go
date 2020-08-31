package devices

import (
	"net/http"

	auth "github.com/soapboxsocial/soapbox/pkg/api/middleware"
	"github.com/soapboxsocial/soapbox/pkg/devices"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
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
	if err != nil && err.Error() != "pq: duplicate key value violates unique constraint \"devices_pkey\"" {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToStoreDevice, "failed")
		return
	}

	httputil.JsonSuccess(w)
}
