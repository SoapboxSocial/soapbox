package main

// This command is responsible for registering welcome rooms when a new user signs up.

import (
	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	queue := pubsub.NewQueue(rdb)

	events := queue.Subscribe(pubsub.UserTopic)

	for evt := range events {
		if evt.Type != pubsub.EventTypeNewUser {
			continue
		}
	}
}
