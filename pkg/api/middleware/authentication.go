package middleware

import (
	"context"
	"net/http"

	httputil "github.com/ephemeral-networks/voicely/pkg/http"
	"github.com/ephemeral-networks/voicely/pkg/sessions"
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
			httputil.JsonError(w, 401, httputil.ErrorCodeUnauthorized, "unauthorized")
			return
		}

		id, err := h.sm.GetUserIDForSession(token)
		if err != nil || id == 0 {
			httputil.JsonError(w, 401, httputil.ErrorCodeUnauthorized, "unauthorized")
			return
		}

		ctx := context.WithValue(req.Context(), "id", id)
		r := req.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
