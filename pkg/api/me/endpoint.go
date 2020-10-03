package me

import (
	"log"
	"net/http"

	auth "github.com/soapboxsocial/soapbox/pkg/api/middleware"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type MeEndpoint struct {
	users *users.UserBackend
	ns    *notifications.Storage
}

func NewMeEndpoint(users *users.UserBackend, ns *notifications.Storage) *MeEndpoint {
	return &MeEndpoint{
		users: users,
		ns:    ns,
	}
}

func (m *MeEndpoint) GetMe(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	user, err := m.users.FindByID(id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetUser, "failed to get self")
		return
	}

	err = httputil.JsonEncode(w, user)
	if err != nil {
		log.Printf("failed to write me response: %s\n", err.Error())
	}
}

func (m *MeEndpoint) GetNotifications(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	list, err := m.ns.GetNotifications(id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToGetUser, "failed to get self")
		return
	}

	err = httputil.JsonEncode(w, list)
	if err != nil {
		log.Printf("failed to write me response: %s\n", err.Error())
	}
}
