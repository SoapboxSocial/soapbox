package middlewares

import (
	"net/http"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
)

type authenticationHandler struct {
	sm *sessions.SessionManager
}

func NewAuthenticationMiddleware(sm *sessions.SessionManager) *authenticationHandler {
	return &authenticationHandler{
		sm: sm,
	}
}

func (h authenticationHandler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		token := req.Header.Get("Authorization")
		if token == "" {
			httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeUnauthorized, "unauthorized")
			return
		}

		id, err := h.sm.GetUserIDForSession(token)
		if err != nil || id == 0 {
			httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeUnauthorized, "unauthorized")
			return
		}

		r := req.WithContext(httputil.WithUserID(req.Context(), id))

		next.ServeHTTP(w, r)
	})
}
