package cmd

import (
	"log"

	"github.com/dghubble/oauth1"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/soapboxsocial/soapbox/pkg/linkedaccounts"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows/providers"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows/worker"
	"github.com/soapboxsocial/soapbox/pkg/redis"
	"github.com/soapboxsocial/soapbox/pkg/sql"
)

var followscmd = &cobra.Command{
	Use:   "follows",
	Short: "created recommendations for users to follow",
	RunE:  runFollows,
}

func runFollows(*cobra.Command, []string) error {
	rdb := redis.NewRedis(config.Redis)

	db, err := sql.Open(config.DB)
	if err != nil {
		return errors.Wrap(err, "failed to open db")
	}

	oauth := oauth1.NewConfig(
		config.Twitter.Key,
		config.Twitter.Secret,
	)

	dispatch := worker.NewDispatcher(3, &worker.Config{
		Twitter:         providers.NewTwitter(oauth, linkedaccounts.NewLinkedAccountsBackend(db), nil),
		Recommendations: follows.NewBackend(db),
		Queue:           pubsub.NewQueue(rdb),
	})
	dispatch.Run()

	row := db.QueryRow("SELECT COUNT(user_id) FROM linked_accounts")

	var count int
	err = row.Scan(&count)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("found %d accounts\n", count)

	rows, err := db.Query("SELECT user_id FROM linked_accounts")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			log.Println(err)
			continue
		}

		dispatch.Dispatch(id)
	}

	dispatch.Wait()
	dispatch.Stop()

	return nil
}
