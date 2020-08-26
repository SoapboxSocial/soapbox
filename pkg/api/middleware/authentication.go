package middleware

import (
	"context"
	"net/http"

	httputil "github.com/ephemeral-networks/soapbox/pkg/http"
	"github.com/ephemeral-networks/soapbox/pkg/sessions"
)

type key string

const userID key = "id"

func GetUserIDFromContext(ctx context.Context) (int, bool) {
	val := ctx.Value(userID)
	id, ok := val.(int)
	return id, ok
}

func WithUserID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, userID, id)
}

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

		r := req.WithContext(WithUserID(req.Context(), id))

		next.ServeHTTP(w, r)
	})
}
