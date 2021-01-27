package internal

import (
	"errors"
	"net/http"
)

// SessionID returns the authorization key stored in the request.
func SessionID(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", errors.New("no authorization token")
	}

	return token, nil
}
