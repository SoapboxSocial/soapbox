package apple

import (
	"context"
	"fmt"
	"log"
	"strings"

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

func NewSignInWithAppleAppValidation(client *apple.Client, teamID, clientID, keyID, key string) (*SignInWithAppleAppValidation, error) {
	secret, err := apple.GenerateClientSecret(key, teamID, clientID, keyID)
	if err != nil {
		return nil, fmt.Errorf("apple.GenerateClientSecret err: %v", err)
	}

	return &SignInWithAppleAppValidation{
		client: client,
		id:     clientID,
		secret: secret,
	}, nil
}

// Validate validates a JWT token and returns UserInfo
func (s *SignInWithAppleAppValidation) Validate(jwt string) (*UserInfo, error) {
	req := apple.AppValidationTokenRequest{
		ClientID:     s.id,
		ClientSecret: s.secret,
		Code:         jwt,
	}

	resp := apple.ValidationResponse{}

	err := s.client.VerifyAppToken(context.Background(), req, &resp)
	if err != nil {
		return nil, fmt.Errorf("s.client.VerifyAppToken err: %v", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("apple response err: %v", resp.Error)
	}

	log.Printf("%v", resp)

	userID, err := apple.GetUniqueID(resp.IDToken)
	if err != nil {
		return nil, fmt.Errorf("apple.GetUniqueID err: %v", err)
	}

	claim, err := apple.GetClaims(resp.IDToken)
	if err != nil {
		return nil, fmt.Errorf("apple.GetClaims err: %v", err)
	}

	email, ok := claim.GetString("email")
	if !ok {
		return nil, fmt.Errorf("claim.GetString err: %v", err)
	}

	return &UserInfo{Email: strings.ToLower(email), ID: userID}, nil
}
