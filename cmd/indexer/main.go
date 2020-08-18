package main

import (
	"log"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/ephemeral-networks/voicely/pkg/indexer"
)

type handlerFunc func(*indexer.Event) error

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	queue := indexer.NewIndexerQueue(rdb)

	for {
		if queue.Len() == 0 {
			// @todo think about this timeout
			time.Sleep(1 * time.Second)
			continue
		}

		event, err := queue.Pop()
		if err != nil {
			log.Printf("failed to pop from queue: %s\n", err)
			continue
		}

		go handleEvent(event)
	}
}

func handleEvent(event *indexer.Event) {
	handler := getHandler(event.Type)
	if handler == nil {
		log.Printf("no event handler for type \"%d\"\n", event.Type)
		return
	}

	err := handler(event)
	if err != nil {
		log.Printf("handler \"%d\" failed with error: %s\n", event.Type, err.Error())
	}
}

func getHandler(eventType indexer.EventType) handlerFunc {
	switch eventType {
	case indexer.EventTypeUserUpdate:
		return handleUserUpdate
	default:
		return nil
	}
}

func handleUserUpdate(event *indexer.Event) error {
	return nil
}