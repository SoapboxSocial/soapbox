package devices

import (
	"log"
	"net/http"

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
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "")
		return
	}

	token := r.Form.Get("token")
	if token == "" {
		httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid token")
		return
	}

	err = d.db.AddDeviceForUser(r.Context().Value("id").(int), token)
	if err != nil {
		log.Println("failed to create session: ", err.Error())
		httputil.JsonError(w, 500, httputil.ErrorCodeFailedToStoreDevice, "failed")
		return
	}

	httputil.JsonSuccess(w)
}
