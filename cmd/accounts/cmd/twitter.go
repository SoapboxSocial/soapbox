package cmd

import (
	"log"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/soapboxsocial/soapbox/pkg/sql"
)

const errExpiredTokenCode = 89

var twitterCmd = &cobra.Command{
	Use:   "twitter",
	Short: "twitter account management",
	RunE:  runTwitter,
}

func runTwitter(*cobra.Command, []string) error {
	db, err := sql.Open(config.DB)
	if err != nil {
		return errors.Wrap(err, "failed to open db")
	}

	oauth := oauth1.NewConfig(
		config.Twitter.Key,
		config.Twitter.Secret,
	)

	return nil
}

func update(oauth oauth1.Config, token, secret string) {
	access := oauth1.NewToken(token, secret)
	httpClient := oauth.Client(oauth1.NoContext, access)

	client := twitter.NewClient(httpClient)
	user, _, err := client.Accounts.VerifyCredentials(nil)
	if err != nil {
		aerr, ok := err.(twitter.APIError)
		if !ok {
			return
		}

		if aerr.Contains(errExpiredTokenCode) {
			// @TODO
		}

		log.Printf("accounts.VerifyCredentials failed err %s\n", err)
	}
}
