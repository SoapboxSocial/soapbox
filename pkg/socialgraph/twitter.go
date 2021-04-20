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

	client := twitter.NewClient(httpClient)
	//client.Friendships.()

	return nil, nil
}
