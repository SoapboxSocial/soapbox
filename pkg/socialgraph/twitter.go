package socialgraph

import (
	"errors"
	"log"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"

	"github.com/soapboxsocial/soapbox/pkg/linkedaccounts"
)

type Twitter struct {
	oauth oauth1.Config

	backend linkedaccounts.Backend
}

func (t *Twitter) FindPeopleToFollowFor(user int) ([]int, error) {
	account, err := t.backend.GetTwitterProfileFor(user)
	if err != nil {
		return nil, err
	}

	access := oauth1.NewToken(account.Token, account.Secret)
	httpClient := t.oauth.Client(oauth1.NoContext, access)

	accounts, err := t.backend.GetAllTwitterProfilesForUsersNotFollowedBy(user)
	if err != nil {
		return nil, err
	}

	parts := chunkAccounts(accounts, 100)

	client := twitter.NewClient(httpClient)

	friendships := make([]twitter.FriendshipResponse, 0)
	for _, part := range parts {
		resp, err := request(client, part)
		if err != nil {
			log.Printf("request err: %s\n", err)
			continue
		}

		friendships = append(friendships, resp...)
	}

	ids := make([]int, 0)
	for _, account := range accounts {
		if isFollowedOnTwitter(account, friendships) {
			ids = append(ids, account.ID)
		}
	}

	return ids, nil
}

func isFollowedOnTwitter(account linkedaccounts.LinkedAccount, friendships []twitter.FriendshipResponse) bool {

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
