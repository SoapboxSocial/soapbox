package rooms

import (
	"context"
	"errors"

	"google.golang.org/grpc/metadata"
)

// SessionID returns the authorization key stored in the grpc metadata.
func SessionID(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}

	auth := md.Get("authorization")
	if len(auth) != 1 {
		return "", errors.New("unauthorized")
	}

	return auth[0], nil
}
