package apple

import (
	"context"
	"errors"

	"github.com/Timothylock/go-signin-with-apple/apple"
)

// UserInfo contains the apple ID user info
type UserInfo struct {
	ID    string
	Email string
}

// SignInWithApple defines an interface for validating apple sign ins
type SignInWithApple interface {
	Validate(jwt string) (*UserInfo, error)
}

type SignInWithAppleAppValidation struct {
	client *apple.Client

	id     string
	secret string
}

func NewSignInWithAppleAppValidation(client *apple.Client, id, secret string) *SignInWithAppleAppValidation {
	return &SignInWithAppleAppValidation{
		client: client,
		id:     id,
		secret: secret,
	}
}

// Validate validates a JWT token and returns UserInfo
func (s *SignInWithAppleAppValidation) Validate(jwt string) (*UserInfo, error) {
	req := apple.AppValidationTokenRequest{
		ClientID:     s.id,
		ClientSecret: s.secret,
		Code:         jwt,
	}

	resp := &apple.ValidationResponse{}

	err := s.client.VerifyAppToken(context.Background(), req, resp)
	if err != nil {
		return nil, err
	}

	userID, err := apple.GetUniqueID(resp.IDToken)
	if err != nil {
		return nil, err
	}

	claim, err := apple.GetClaims(resp.IDToken)
	if err != nil {
		return nil, err
	}

	email, ok := claim.GetString("email")
	if !ok {
		return nil, errors.New("failed to recover email")
	}

	return &UserInfo{Email: email, ID: userID}, nil
}
