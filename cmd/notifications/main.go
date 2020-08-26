package main

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"

	"github.com/ephemeral-networks/soapbox/pkg/devices"
	"github.com/ephemeral-networks/soapbox/pkg/notifications"
	"github.com/ephemeral-networks/soapbox/pkg/users"
)

var devicesBackend *devices.DevicesBackend
var userBackend *users.UserBackend
var service *notifications.Service

type handlerFunc func(*notifications.Event) ([]string, *notifications.Notification, error)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	queue := notifications.NewNotificationQueue(rdb)

	db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=voicely password=voicely dbname=voicely sslmode=disable")
	if err != nil {
		panic(err)
	}

	devicesBackend = devices.NewDevicesBackend(db)
	userBackend = users.NewUserBackend(db)

	authKey, err := token.AuthKeyFromFile("/conf/authkey.p8")
	if err != nil {
		panic(err)
	}

	// @todo add flag for which enviroment

	client := apns2.NewTokenClient(&token.Token{
		AuthKey: authKey,
		KeyID:   "9U8K3MKG2K", // @todo these should be in config files
		TeamID:  "Z9LC5GZ33U",
	}).Production()

	service = notifications.NewService("app.social.soapbox", client)

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

func handleEvent(event *notifications.Event) {
	handler := getHandler(event.Type)
	if handler == nil {
		log.Printf("no event handler for type \"%d\"\n", event.Type)
		return
	}

	targets, notification, err := handler(event)
	if err != nil {
		log.Printf("handler \"%d\" failed with error: %s\n", event.Type, err.Error())
	}

	if notification == nil {
		log.Println("notification unexpectedly nil")
		return
	}

	for _, target := range targets {
		err := service.Send(target, *notification)
		if err != nil {
			log.Printf("failed to send to target \"%s\" with error: %s\n", target, err.Error())
		}
	}
}

func getHandler(eventType notifications.EventType) handlerFunc {
	switch eventType {
	case notifications.EventTypeRoomCreation:
		return onRoomCreation
	case notifications.EventTypeNewFollower:
		return onNewFollower
	default:
		return nil
	}
}

func onRoomCreation(event *notifications.Event) ([]string, *notifications.Notification, error) {
	targets, err := devicesBackend.FetchAllFollowerDevices(event.Creator)
	if err != nil {
		return nil, nil, err
	}

	name := event.Params["name"].(string)
	room, ok := event.Params["id"].(float64)
	if !ok {
		return nil, nil, errors.New("failed to recover room ID")
	}

	displayName, err := getDisplayName(event.Creator)
	if err != nil {
		return nil, nil, err
	}

	notification := func() *notifications.Notification {
		if name == "" {
			return notifications.NewRoomNotification(int(room), displayName)
		}

		return notifications.NewRoomNotificationWithName(int(room), displayName, name)
	}()

	return targets, notification, nil
}

func onNewFollower(event *notifications.Event) ([]string, *notifications.Notification, error) {
	targetID, ok := event.Params["id"].(float64)
	if !ok {
		return nil, nil, errors.New("failed to recover target ID")
	}

	targets, err := devicesBackend.GetDevicesForUser(int(targetID))
	if err != nil {
		return nil, nil, err
	}

	displayName, err := getDisplayName(event.Creator)
	if err != nil {
		return nil, nil, err
	}

	return targets, notifications.NewFollowerNotification(event.Creator, displayName), nil
}

func getDisplayName(id int) (string, error) {
	user, err := userBackend.FindByID(id)
	if err != nil {
		return "", err
	}

	return user.DisplayName, nil
}
