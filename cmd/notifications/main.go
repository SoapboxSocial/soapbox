package main

import (
	"log"

	"github.com/go-redis/redis/v8"

	"github.com/ephemeral-networks/voicely/pkg/notifications"
)

func main() {

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	queue := notifications.NewNotificationQueue(rdb)

	for {
		event, err := queue.Pop()
		if err != nil {
			log.Print(err)
			continue
		}

		go handleEvent(event)
	}
}

func handleEvent(event *notifications.Event) {
	switch event.Type {
	case notifications.EventTypeRoomCreation:
		onRoomCreation(event)
	default:
		log.Printf("unknown event type \"%d\"", event.Type)
	}
}

func onRoomCreation(event *notifications.Event) {
	// @todo gather all followers with devices
}
