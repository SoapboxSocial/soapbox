package cmd

import (
	"context"
	"log"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/linkedaccounts"
	"github.com/soapboxsocial/soapbox/pkg/notifications/pb"
	"github.com/soapboxsocial/soapbox/pkg/sql"
)

const errExpiredTokenCode = 89

var (
	accounts      *linkedaccounts.Backend
	notifications pb.NotificationServiceClient

	addr string

	twitterCmd = &cobra.Command{
		Use:   "twitter",
		Short: "twitter account management used for deleting and updating out-of-date twitter accounts",
		RunE:  runTwitter,
	}
)

func init() {
	twitterCmd.Flags().StringVarP(&addr, "addr", "a", "127.0.0.1:50053", "grpc address")
}

func runTwitter(*cobra.Command, []string) error {
	db, err := sql.Open(config.DB)
	if err != nil {
		return errors.Wrap(err, "failed to open db")
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	defer conn.Close()

	notifications = pb.NewNotificationServiceClient(conn)

	oauth := oauth1.NewConfig(
		config.Twitter.Key,
		config.Twitter.Secret,
	)

	accounts = linkedaccounts.NewLinkedAccountsBackend(db)

	rows, err := db.Query("SELECT user_id, token, secret FROM linked_accounts")
	if err != nil {
		return err
	}

	for rows.Next() {
		var (
			id     int
			token  string
			secret string
		)

		err := rows.Scan(&id, &token, &secret)
		if err != nil {
			log.Println(err)
			continue
		}

		update(oauth, id, token, secret)
	}

	return nil
}

func update(oauth *oauth1.Config, user int, token, secret string) {
	access := oauth1.NewToken(token, secret)
	httpClient := oauth.Client(oauth1.NoContext, access)

	client := twitter.NewClient(httpClient)
	profile, _, err := client.Accounts.VerifyCredentials(nil)
	if err != nil {
		aerr, ok := err.(twitter.APIError)
		if !ok {
			return
		}

		if aerr.Contains(errExpiredTokenCode) {
			unlink(user)
			return
		}

		log.Printf("accounts.VerifyCredentials failed err %s\n", err)
	}

	err = accounts.UpdateTwitterUsernameFor(user, profile.ScreenName)
	if err != nil {
		log.Printf("accounts.UpdateTwitterUsernameFor err: %s", err)
	}
}

func unlink(user int) {
	log.Printf("removing twitter for %d\n", user)

	err := accounts.UnlinkTwitterProfile(user)
	if err != nil {
		log.Printf("accounts.UnlinkTwitterProfile err %s\n", err)
	}

	_, err = notifications.SendNotification(context.Background(), &pb.SendNotificationRequest{
		Targets: []int64{int64(user)},
		Notification: &pb.Notification{
			Category: "INFO",
			Alert: &pb.Notification_Alert{
				Body: "There was an issue with your twitter account. Please reconnect it.",
			},
		},
	})

	if err != nil {
		log.Printf("notifications.SendNotification err: %s\n", err)
	}
}
