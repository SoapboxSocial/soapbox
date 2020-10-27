package groups

import (
	"log"
	"net/http"
	"strings"

	auth "github.com/soapboxsocial/soapbox/pkg/api/middleware"
	"github.com/soapboxsocial/soapbox/pkg/groups"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
)

type Endpoint struct {
	backend *groups.Backend
}

func NewEndpoint(backend *groups.Backend) *Endpoint {
	return &Endpoint{
		backend: backend,
	}
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

	bio := strings.TrimSpace(strings.ReplaceAll(r.Form.Get("bio"), "\n", " "))
	if len([]rune(bio)) > 300 {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "bio too long")
		return
	}

	// @todo bio

	id, err := e.backend.CreateGroup(userID, name, bio, "", "")
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "failed to create")
		return
	}

	err = httputil.JsonEncode(w, map[string]interface{}{"success": true, "id": id})
	if err != nil {
		log.Println("error writing response: " + err.Error())
	}
}
