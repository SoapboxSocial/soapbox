package main

import (
	"log"
	"strconv"

	"github.com/dukex/mixpanel"
	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/tracking"
)


func main() {
	client := mixpanel.New("d124ce8f1516eb7baa7980f4de68ded5", "https://api-eu.mixpanel.com")
	tracker := tracking.NewMixpanelTracker(client)

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
			err := tracker.Track(event)
			if err != nil {
				log.Printf("tracker.Track err %v", err)
			}
		}()
	}
}

func handleEvent(event *pubsub.Event) *tracking.Event {
	switch event.Type {
	case pubsub.EventTypeNewRoom:
		id, err := event.GetInt("creator")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "room_new",
			Properties: map[string]interface{}{
				"room_id":    event.Params["id"],
				"visibility": event.Params["visibility"],
			},
		}
	case pubsub.EventTypeRoomJoin:
		id, err := event.GetInt("creator")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "room_join",
			Properties: map[string]interface{}{
				"room_id":    event.Params["id"],
				"visibility": event.Params["visibility"],
			},
		}
	case pubsub.EventTypeRoomLeft:
		id, err := event.GetInt("creator")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "room_left",
			Properties: map[string]interface{}{
				"room_id": event.Params["id"],
			},
		}
	case pubsub.EventTypeNewUser:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "new_user",
			Properties: map[string]interface{}{
				"user_id":  event.Params["id"],
				"username": event.Params["username"],
			},
		}
	case pubsub.EventTypeNewGroup:
		id, err := event.GetInt("creator")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "group_new",
			Properties: map[string]interface{}{
				"group_id": event.Params["id"],
				"name":     event.Params["name"],
			},
		}
	case pubsub.EventTypeNewGroupRoom:
		id, err := event.GetInt("creator")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "room_new",
			Properties: map[string]interface{}{
				"room_id":    event.Params["id"],
				"visibility": event.Params["visibility"],
				"group_id":   event.Params["group"],
			},
		}
	case pubsub.EventTypeNewFollower:
		id, err := event.GetInt("follower")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "followed",
			Properties: map[string]interface{}{
				"following_id": event.Params["id"],
			},
		}
	case pubsub.EventTypeGroupJoin:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "group_join",
			Properties: map[string]interface{}{
				"group": event.Params["group"],
			},
		}
	case pubsub.EventTypeNewStory:
		id, err := event.GetInt("creator")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:         strconv.Itoa(id),
			Name:       "story_new",
			Properties: map[string]interface{}{},
		}
	case pubsub.EventTypeStoryReaction:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:         strconv.Itoa(id),
			Name:       "story_reaction",
			Properties: map[string]interface{}{},
		}
	case pubsub.EventTypeUserHeartbeat:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:         strconv.Itoa(id),
			Name:       "heartbeat",
			Properties: map[string]interface{}{},
		}
	}

	return nil
}
