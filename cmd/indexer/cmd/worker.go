package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/spf13/cobra"

	"github.com/soapboxsocial/soapbox/pkg/groups"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/redis"
	"github.com/soapboxsocial/soapbox/pkg/sql"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

var worker = &cobra.Command{
	Use:   "worker",
	Short: "runs a index worker",
	RunE:  runWorker,
}

var errNoRequestHandler = errors.New("no request handler for event")

func runWorker(*cobra.Command, []string) error {
	rdb := redis.NewRedis(config.Redis)

	db, err := sql.Open(config.DB)
	if err != nil {
		return err
	}

	client, err = elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	queue := pubsub.NewQueue(rdb)
	events := queue.Subscribe(pubsub.UserTopic, pubsub.GroupTopic)

	userBackend = users.NewUserBackend(db)
	groupsBackend = groups.NewBackend(db)

	for event := range events {
		go handleEvent(event)
	}

	return nil
}

func handleEvent(event *pubsub.Event) {
	request, err := requestFor(event)
	if err != nil {
		if err == errNoRequestHandler {
			return
		}

		log.Printf("failed to create request: %v\n", err)
		return
	}

	res, err := request.Do(context.Background(), client)
	if err != nil {
		log.Printf("failed to execute request: %v\n", err)
	}

	_ = res.Body.Close()
}

func requestFor(event *pubsub.Event) (esapi.Request, error) {
	switch event.Type {
	case pubsub.EventTypeUserUpdate, pubsub.EventTypeNewUser, pubsub.EventTypeNewFollower: // @TODO think about unfollows
		return userUpdateRequest(event)
	case pubsub.EventTypeNewGroup, pubsub.EventTypeGroupUpdate:
		return groupUpdateRequest(event)
	case pubsub.EventTypeGroupDelete:
		return groupDeleteRequest(event)
	default:
		return nil, errNoRequestHandler
	}
}

func groupUpdateRequest(event *pubsub.Event) (esapi.Request, error) {
	id, ok := event.Params["id"].(float64)
	if !ok {
		return nil, errors.New("failed to recover user ID")
	}

	group, err := groupsBackend.FindById(int(id))
	if err != nil {
		return nil, err
	}

	if group.GroupType == "private" {
		return nil, errors.New("private group cannot be indexed")
	}

	body, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}

	return esapi.IndexRequest{
		Index:      "groups",
		DocumentID: strconv.Itoa(group.ID),
		Body:       strings.NewReader(string(body)),
		Refresh:    "true",
	}, nil
}

func userUpdateRequest(event *pubsub.Event) (esapi.Request, error) {
	id, ok := event.Params["id"].(float64)
	if !ok {
		return nil, errors.New("failed to recover user ID")
	}

	user, err := userBackend.GetUserForSearchEngine(int(id))
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	return esapi.IndexRequest{
		Index:      "users",
		DocumentID: strconv.Itoa(user.ID),
		Body:       strings.NewReader(string(body)),
		Refresh:    "true",
	}, nil
}

func groupDeleteRequest(event *pubsub.Event) (esapi.Request, error) {
	id, ok := event.Params["group"].(float64)
	if !ok {
		return nil, errors.New("failed to recover group ID")
	}

	return esapi.DeleteRequest{
		Index:      "groups",
		DocumentID: strconv.Itoa(int(id)),
		Refresh:    "true",
	}, nil
}
