package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/dukex/mixpanel"

	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/redis"
	"github.com/soapboxsocial/soapbox/pkg/tracking"
)

type Conf struct {
	Mixpanel struct {
		Token string `mapstructure:"token"`
		URL   string `mapstructure:"url"`
	} `mapstructure:"mixpanel"`
	Redis conf.RedisConf `mapstructure:"redis"`
}

func parse() (*Conf, error) {
	var file string
	flag.StringVar(&file, "c", "config.toml", "config file")
	flag.Parse()

	config := &Conf{}
	err := conf.Load(file, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func main() {
	config, err := parse()
	if err != nil {
		log.Fatal("failed to parse config")
	}

	client := mixpanel.New(config.Mixpanel.Token, config.Mixpanel.URL)
	tracker := tracking.NewMixpanelTracker(client)

	rdb := redis.NewRedis(config.Redis)
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
			Name: tracking.NewUser,
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

		// @TODO ADD HEARTBEAT COOLDOWN of 30 mins

		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:         strconv.Itoa(id),
			Name:       "heartbeat",
			Properties: map[string]interface{}{},
		}
	case pubsub.EventTypeRoomOpenMini:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "room_open_mini",
			Properties: map[string]interface{}{
				"room_id": event.Params["room"],
				"mini":    event.Params["mini"],
			},
		}
	case pubsub.EventTypeRoomLinkShare:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "room_link_share",
			Properties: map[string]interface{}{
				"room_id": event.Params["room"],
			},
		}
	case pubsub.EventTypeDeleteUser:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: tracking.DeleteUser,
		}
	}

	return nil
}
