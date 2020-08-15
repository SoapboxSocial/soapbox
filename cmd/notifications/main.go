package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"

	"github.com/ephemeral-networks/voicely/pkg/devices"
	"github.com/ephemeral-networks/voicely/pkg/notifications"
	"github.com/ephemeral-networks/voicely/pkg/users"
)

var devicesBackend *devices.DevicesBackend
var userBackend *users.UserBackend
var service *notifications.Service

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
		KeyID:   "82439YH93F",
		TeamID:  "Z9LC5GZ33U",
	}).Production()

	service = notifications.NewService("com.voicely.voicely", client)

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

	err := handler(event)
	if err != nil {
		log.Printf("handler \"%d\" failed with error: %s", event.Type, err.Error())
	}
}

func getHandler(eventType notifications.EventType) func(*notifications.Event) error {
	switch eventType {
	case notifications.EventTypeRoomCreation:
		return onRoomCreation
	default:
		return nil
	}
}

func onRoomCreation(event *notifications.Event) error {
	targets, err := devicesBackend.FetchAllFollowerDevices(event.Creator)
	if err != nil {
		return err
	}

	name := event.Params["name"].(string)
	room := int(event.Params["id"].(float64))

	displayName, err := getDisplayName(event.Creator)
	if err != nil {
		return err
	}

	var notification notifications.Notification
	if name == "" {
		notification = notifications.NewRoomNotification(room, displayName)
	} else {
		notification = notifications.NewRoomNotificationWithName(room, displayName, name)
	}

	for _, target := range targets {
		err = service.Send(target, notification)
		if err != nil {
			log.Printf("failed to send to target \"%s\" with error: %s", target, err.Error())
		}
	}

	return nil
}

func getDisplayName(id int) (string, error) {
	user, err := userBackend.FindByID(id)
	if err != nil {
		return "", err
	}

	return user.DisplayName, nil
}
