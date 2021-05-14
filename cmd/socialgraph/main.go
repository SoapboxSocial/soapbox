package main

import (
	"log"

	"github.com/dghubble/oauth1"

	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/linkedaccounts"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows/providers"
	"github.com/soapboxsocial/soapbox/pkg/sql"
)

func main() {
	db, err := sql.Open(conf.PostgresConf{})

	if err != nil {
		log.Fatal(err)
	}

	oauth := oauth1.NewConfig(
		"nAzgMi6loUf3cl0hIkkXhZSth",
		"sFQEQ2cjJZSJgepUMmNyeTxiGggFXA1EKfSYAXpbARTu3CXBQY",
	)

	twitter := providers.NewTwitter(oauth, linkedaccounts.NewLinkedAccountsBackend(db))

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

		found, err := twitter.FindUsersToFollowFor(id)
		if err != nil {
			log.Printf("find err: %s for %d", err, id)
			continue
		}

		log.Printf("Found %d for %d\n", len(found), id)
	}

}
