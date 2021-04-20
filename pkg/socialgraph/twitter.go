package socialgraph

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"

	"github.com/soapboxsocial/soapbox/pkg/linkedaccounts"
)

type Twitter struct {
	oauth oauth1.Config

	backend linkedaccounts.Backend
}

func (t *Twitter) FindFriendsFor(user int) ([]int, error) {
	account, err := t.backend.GetTwitterProfileFor(user)
	if err != nil {
		return nil, err
	}

	access := oauth1.NewToken(account.Token, account.Secret)
	httpClient := t.oauth.Client(oauth1.NoContext, access)

	// @TODO paginate

	accounts, err := t.backend.GetAllTwitterProfiles(user)
	if err != nil {
		return nil, err
	}

	parts := chunkAccounts(accounts)

	client := twitter.NewClient(httpClient)

	// @TODO DO THE REQUESTS, probably go func it with a wait group

	res, _, err := client.Friendships.Lookup(nil) // @Todo
	if err != nil {
		return nil, err
	}

	for _, resp := range res {

	}

	//client.Friendships.()

	return nil, nil
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
