package providers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"golang.org/x/sync/errgroup"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"

	"github.com/soapboxsocial/soapbox/pkg/linkedaccounts"
)

type Twitter struct {
	oauth *oauth1.Config

	transport *http.Client

	backend *linkedaccounts.Backend
}

func NewTwitter(oauth *oauth1.Config, backend *linkedaccounts.Backend, transport *http.Client) *Twitter {
	return &Twitter{
		oauth:     oauth,
		backend:   backend,
		transport: transport,
	}
}

func (t *Twitter) FindUsersToFollowFor(user int) ([]int, error) {
	client, err := t.getClientForUser(user)
	if err != nil {
		return nil, err
	}

	accounts, err := t.backend.GetAllTwitterProfilesForUsersNotRecommendedToAndNotFollowedBy(user)
	if err != nil {
		return nil, err
	}

	parts := chunkAccounts(accounts, 100)

	friendships := make([]twitter.FriendshipResponse, 0)

	var wg errgroup.Group
	for _, part := range parts {
		accounts := part
		wg.Go(func() error {
			resp, err := request(client, accounts)
			if err != nil { // @TODO check if the error is twitter error with old account, if yes we delete.
				return err
			}

			friendships = append(friendships, resp...)
			return nil
		})
	}

	err = wg.Wait()
	if err != nil {
		log.Printf("request err: %s\n", err)
	}

	ids := make([]int, 0)
	for _, account := range accounts {
		if isFollowedOnTwitter(account, friendships) {
			ids = append(ids, account.ID)
		}
	}

	return ids, nil
}

func (t *Twitter) getClientForUser(id int) (*twitter.Client, error) {
	account, err := t.backend.GetTwitterProfileFor(id)
	if err != nil {
		return nil, err
	}

	access := oauth1.NewToken(account.Token, account.Secret)

	ctx := oauth1.NoContext
	if t.transport != nil {
		ctx = context.WithValue(ctx, oauth1.HTTPClient, t.transport)
	}

	httpClient := t.oauth.Client(ctx, access)
	return twitter.NewClient(httpClient), nil
}

func isFollowedOnTwitter(account linkedaccounts.LinkedAccount, friendships []twitter.FriendshipResponse) bool {
	for _, friendship := range friendships {
		if friendship.ID != account.ProfileID {
			continue
		}

		for _, conn := range friendship.Connections {
			if conn == "following" {
				return true
			}
		}
	}

	return false
}

func request(client *twitter.Client, accounts []linkedaccounts.LinkedAccount) ([]twitter.FriendshipResponse, error) {
	ids := make([]int64, 0)
	for _, account := range accounts {
		ids = append(ids, account.ProfileID)
	}

	res, _, err := client.Friendships.Lookup(&twitter.FriendshipLookupParams{UserID: ids}) // @Todo
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, errors.New("no response")
	}

	return *res, nil
}

func chunkAccounts(accounts []linkedaccounts.LinkedAccount, chunkSize int) [][]linkedaccounts.LinkedAccount {
	var divided [][]linkedaccounts.LinkedAccount

	for i := 0; i < len(accounts); i += chunkSize {
		end := i + chunkSize

		if end > len(accounts) {
			end = len(accounts)
		}

		divided = append(divided, accounts[i:end])
	}

	return divided
}
