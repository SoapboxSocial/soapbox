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

	accounts, err := t.backend.GetAllTwitterProfilesForUsersNotFollowedBy(user)
	if err != nil {
		return nil, err
	}

	parts := chunkAccounts(accounts, 100)

	client := twitter.NewClient(httpClient)

	for _, part := range parts {
		_ = request(client, part)
	}

	return nil, nil
}

func request(client *twitter.Client, accounts []linkedaccounts.LinkedAccount) []linkedaccounts.LinkedAccount {
	ids := make([]int64, 0)
	for _, account := range accounts {
		ids = append(ids, account.ProfileID)
	}

	res, _, err := client.Friendships.Lookup(&twitter.FriendshipLookupParams{UserID: ids}) // @Todo
	if err != nil {
		return nil
	}

	ret := make([]linkedaccounts.LinkedAccount, 0)

	OUTER:
	for _, account := range accounts {
		for _, resp := range *res {
			if resp.ID == account.ProfileID {
				continue OUTER
			}
		}

		ret = append(ret, account)

	}

	return ret
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
