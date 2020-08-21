package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/go-redis/redis/v8"

	"github.com/ephemeral-networks/voicely/pkg/indexer"
	"github.com/ephemeral-networks/voicely/pkg/users"
)

type handlerFunc func(*indexer.Event) error

var client *elasticsearch.Client
var userBackend *users.UserBackend

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	queue := indexer.NewIndexerQueue(rdb)

	db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=voicely password=voicely dbname=voicely sslmode=disable")
	if err != nil {
		panic(err)
	}

	client, err = elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	userBackend = users.NewUserBackend(db)

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
	id, ok := event.Params["id"].(float64)
	if !ok {
		return errors.New("failed to recover user ID")
	}

	user, err := userBackend.FindByID(int(id))
	if err != nil {
		return err
	}

	user.Email = nil

	body, err := json.Marshal(user)
	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      "users",
		DocumentID: strconv.Itoa(user.ID),
		Body:       strings.NewReader(string(body)),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), client)
	if err != nil {
		return err
	}

	return res.Body.Close()
}
