package providers

import (
	"errors"
	"log"
	"sync"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"

	"github.com/soapboxsocial/soapbox/pkg/linkedaccounts"
)

type Twitter struct {
	oauth *oauth1.Config

	backend *linkedaccounts.Backend
}

func NewTwitter(oauth *oauth1.Config, backend *linkedaccounts.Backend) *Twitter {
	return &Twitter{
		oauth:   oauth,
		backend: backend,
	}
}

func (t *Twitter) FindUsersToFollowFor(user int) ([]int, error) {
	client, err := t.getClientForUser(user)
	if err != nil {
		return nil, err
	}

	accounts, err := t.backend.GetAllTwitterProfilesForUsersNotFollowedBy(user)
	if err != nil {
		return nil, err
	}

	parts := chunkAccounts(accounts, 100)

	friendships := make([]twitter.FriendshipResponse, 0)

	var wg sync.WaitGroup
	for _, part := range parts {
		wg.Add(1)

		go func(accounts []linkedaccounts.LinkedAccount) {
			defer wg.Done()

			resp, err := request(client, accounts)
			if err != nil {
				log.Printf("request err: %s\n", err)
				return
			}

			friendships = append(friendships, resp...)
		}(part)
	}

	wg.Wait()

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
	httpClient := t.oauth.Client(oauth1.NoContext, access)
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
