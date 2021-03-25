package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"

	"github.com/soapboxsocial/soapbox/pkg/groups"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/redis"
	"github.com/soapboxsocial/soapbox/pkg/sql"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

var writer = &cobra.Command{
	Use:   "writer",
	Short: "runs a index writer, used to reindex",
	RunE:  runWriter,
}

func runWriter(*cobra.Command, []string) error {
	rdb := redis.NewRedis(config.Redis)

	db, err := sql.Open(config.DB)
	if err != nil {
		return err
	}

	queue := pubsub.NewQueue(rdb)
	userBackend = users.NewUserBackend(db)
	groupsBackend = groups.NewBackend(db)

	rows, err := db.Query("SELECT id FROM users;")
	if err != nil {
		return err
	}

	index := 0
	for rows.Next() {
		if index%10 == 0 && index != 0 {
			log.Printf("indexed %d users, sleeping for 5s", index)
			time.Sleep(5 * time.Second)
		}

		var id int
		err := rows.Scan(&id)
		if err != nil {
			log.Printf("error encountered %s", err)
		}

		err = queue.Publish(pubsub.UserTopic, pubsub.NewUserUpdateEvent(id))
		if err != nil {
			log.Printf("error encountered %s", err)
		}

		index++
	}

	log.Printf("finished, total indexed: %d", index)

	return nil
}
