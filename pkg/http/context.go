package http

import "context"

type key string

const userID key = "id"

// GetUserIDFromContext returns a user ID from a context
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	val := ctx.Value(userID)
	id, ok := val.(int)
	return id, ok
}

// WithUserID stores a user ID in the context
func WithUserID(ctx context.Context, id int) context.Context {
	return context.WithValue(ctx, userID, id)
}
