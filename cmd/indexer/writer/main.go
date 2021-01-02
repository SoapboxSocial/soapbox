package main

import (
	"database/sql"

	"github.com/go-redis/redis/v8"

	_ "github.com/lib/pq"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
)

// @TODO CLEAN UP SO WE CAN HAVE A SERVER AND A CLIENT, INDEXING VARIOUS THINGS ETC.

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=voicely password=voicely dbname=voicely sslmode=disable")
	if err != nil {
		panic(err)
	}

	queue := pubsub.NewQueue(rdb)

	rows, err := db.Query("SELECT id FROM users;")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			panic(err)
		}

		err = queue.Publish(pubsub.UserTopic, pubsub.NewUserUpdateEvent(id))
		if err != nil {
			panic(err)
		}
	}
}