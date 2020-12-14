package main

import (
	"errors"
	"log"
	"strconv"

	"github.com/dukex/mixpanel"
	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
)

const new_user = "new_user"

func main() {
	tracker := mixpanel.New("d124ce8f1516eb7baa7980f4de68ded5", "https://api-eu.mixpanel.com")

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	queue := pubsub.NewQueue(rdb)

	events := queue.Subscribe(pubsub.RoomTopic, pubsub.UserTopic, pubsub.GroupTopic, pubsub.StoryTopic)

	for evt := range events {
		event := handleEvent(evt)
		if event == nil {
			continue
		}

		go func() {
			err := tracker.Track(event.id, event.name, &mixpanel.Event{IP: "0", Properties: event.properties})
			if err != nil {
				log.Printf("tracker.Track err: %v\n", err)
			}

			if event.name == new_user {
				err := tracker.Update(event.id, &mixpanel.Update{
					IP:         "0",
					Operation:  "$set",
					Properties: event.properties,
				})

				if err != nil {
					log.Printf("tracker.Update err: %v\n", err)
				}
			}
		}()
	}
}

type Event struct {
	id         string
	name       string
	properties map[string]interface{}
}

func handleEvent(event *pubsub.Event) *Event {
	switch event.Type {
	case pubsub.EventTypeNewRoom:
		id, err := getId(event, "creator")
		if err != nil {
			return nil
		}

		return &Event{
			id:   strconv.Itoa(id),
			name: "room_new",
			properties: map[string]interface{}{
				"room_id":    event.Params["id"],
				"visibility": event.Params["visibility"],
			},
		}
	case pubsub.EventTypeRoomJoin:
		id, err := getId(event, "creator")
		if err != nil {
			return nil
		}

		return &Event{
			id:   strconv.Itoa(id),
			name: "room_join",
			properties: map[string]interface{}{
				"room_id":    event.Params["id"],
				"visibility": event.Params["visibility"],
			},
		}
	case pubsub.EventTypeRoomLeft:
		id, err := getId(event, "creator")
		if err != nil {
			return nil
		}

		return &Event{
			id:   strconv.Itoa(id),
			name: "room_left",
			properties: map[string]interface{}{
				"room_id": event.Params["id"],
			},
		}
	case pubsub.EventTypeNewUser:
		id, err := getId(event, "id")
		if err != nil {
			return nil
		}

		return &Event{
			id:   strconv.Itoa(id),
			name: new_user,
			properties: map[string]interface{}{
				"user_id":  event.Params["id"],
				"username": event.Params["username"],
			},
		}
	case pubsub.EventTypeNewGroup:
		id, err := getId(event, "creator")
		if err != nil {
			return nil
		}

		return &Event{
			id:   strconv.Itoa(id),
			name: "group_new",
			properties: map[string]interface{}{
				"group_id": event.Params["id"],
				"name":     event.Params["name"],
			},
		}
	case pubsub.EventTypeNewGroupRoom:
		id, err := getId(event, "creator")
		if err != nil {
			return nil
		}

		return &Event{
			id:   strconv.Itoa(id),
			name: "room_new",
			properties: map[string]interface{}{
				"room_id":    event.Params["id"],
				"visibility": event.Params["visibility"],
				"group_id":   event.Params["group"],
			},
		}
	case pubsub.EventTypeNewFollower:
		id, err := getId(event, "follower")
		if err != nil {
			return nil
		}

		return &Event{
			id:   strconv.Itoa(id),
			name: "followed",
			properties: map[string]interface{}{
				"following_id": event.Params["id"],
			},
		}
	case pubsub.EventTypeGroupJoin:
		id, err := getId(event, "id")
		if err != nil {
			return nil
		}

		return &Event{
			id:   strconv.Itoa(id),
			name: "group_join",
			properties: map[string]interface{}{
				"group": event.Params["group"],
			},
		}
	case pubsub.EventTypeNewStory:
		id, err := getId(event, "creator")
		if err != nil {
			return nil
		}

		return &Event{
			id:         strconv.Itoa(id),
			name:       "story_new",
			properties: map[string]interface{}{},
		}
	}

	return nil
}

func getId(event *pubsub.Event, field string) (int, error) {
	creator, ok := event.Params[field].(float64)
	if !ok {
		return 0, errors.New("failed to recover creator")
	}

	return int(creator), nil
}
