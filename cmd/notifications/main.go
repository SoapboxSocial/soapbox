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

	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/followers"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/notifications/limiter"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

var (
	errRoomPrivate = errors.New("room is private")
)

var devicesBackend *devices.Backend
var userBackend *users.UserBackend
var followersBackend *followers.FollowersBackend
var service *notifications.Service
var notificationLimiter *limiter.Limiter
var notificationStorage *notifications.Storage

type handlerFunc func(*pubsub.Event) ([]int, *notifications.PushNotification, error)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	queue := pubsub.NewQueue(rdb)

	db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=voicely password=voicely dbname=voicely sslmode=disable")
	if err != nil {
		panic(err)
	}

	devicesBackend = devices.NewBackend(db)
	userBackend = users.NewUserBackend(db)
	followersBackend = followers.NewFollowersBackend(db)
	currentRoom := rooms.NewCurrentRoomBackend(rdb)
	notificationLimiter = limiter.NewLimiter(rdb, currentRoom)
	notificationStorage = notifications.NewStorage(rdb)

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

	events := queue.Subscribe(pubsub.RoomTopic, pubsub.UserTopic)

	for event := range events {
		go handleEvent(event)
	}
}

func handleEvent(event *pubsub.Event) {
	handler := getHandler(event.Type)
	if handler == nil {
		log.Printf("no event handler for type \"%d\"\n", event.Type)
		return
	}

	targets, notification, err := handler(event)
	if err != nil {
		if err == errRoomPrivate {
			return
		}

		log.Printf("handler \"%d\" failed with error: %s\n", event.Type, err.Error())
	}

	if notification == nil {
		log.Println("notification unexpectedly nil")
		return
	}

	for _, target := range targets {
		pushNotification(target, notification)
	}
}

func getHandler(eventType pubsub.EventType) handlerFunc {
	switch eventType {
	case pubsub.EventTypeNewRoom:
		return onRoomCreation
	case pubsub.EventTypeNewFollower:
		return onNewFollower
	case pubsub.EventTypeRoomJoin:
		return onRoomJoined
	case pubsub.EventTypeRoomInvite:
		return onRoomInvite
	default:
		return nil
	}
}

func pushNotification(target int, notification *notifications.PushNotification) {
	if !notificationLimiter.ShouldSendNotification(target, notification.Arguments, notification.Category) {
		return
	}

	d, err := devicesBackend.GetDevicesForUser(target)
	if err != nil {
		log.Printf("devicesBackend.GetDevicesForUser err: %v\n", err)
	}

	for _, device := range d {
		err = service.Send(device, *notification)
		if err != nil {
			log.Printf("failed to send to target \"%s\" with error: %s\n", device, err.Error())
		}
	}

	notificationLimiter.SentNotification(target, notification.Arguments, notification.Category)

	if notification.Category != notifications.NEW_FOLLOWER {
		return
	}

	err = notificationStorage.Store(target, &notifications.Notification{
		Timestamp: time.Now().Second(),
		From:      notification.Arguments["id"].(int),
		Category:  notification.Category,
	})

	if err != nil {
		log.Printf("notificationStorage.Store err: %v\n", err)
	}
}

func onRoomCreation(event *pubsub.Event) ([]int, *notifications.PushNotification, error) {
	if pubsub.RoomVisibility(event.Params["visibility"].(string)) == pubsub.Private {
		return nil, nil, errRoomPrivate
	}

	creator, err := getCreatorId(event)
	if err != nil {
		return nil, nil, err
	}

	targets, err := followersBackend.GetAllFollowerIDsFor(creator)
	if err != nil {
		return nil, nil, err
	}

	name := event.Params["name"].(string)
	room, ok := event.Params["id"].(float64)
	if !ok {
		return nil, nil, errors.New("failed to recover room ID")
	}

	displayName, err := getDisplayName(creator)
	if err != nil {
		return nil, nil, err
	}

	notification := func() *notifications.PushNotification {
		if name == "" {
			return notifications.NewRoomNotification(int(room), displayName)
		}

		return notifications.NewRoomNotificationWithName(int(room), displayName, name)
	}()

	return targets, notification, nil
}

func onRoomJoined(event *pubsub.Event) ([]int, *notifications.PushNotification, error) {
	if pubsub.RoomVisibility(event.Params["visibility"].(string)) == pubsub.Private {
		return nil, nil, errRoomPrivate
	}

	creator, err := getCreatorId(event)
	if err != nil {
		return nil, nil, err
	}

	targets, err := followersBackend.GetAllFollowerIDsFor(creator)
	if err != nil {
		return nil, nil, err
	}

	name := event.Params["name"].(string)
	room, ok := event.Params["id"].(float64)
	if !ok {
		return nil, nil, errors.New("failed to recover room ID")
	}

	displayName, err := getDisplayName(creator)
	if err != nil {
		return nil, nil, err
	}

	notification := func() *notifications.PushNotification {
		if name == "" {
			return notifications.NewRoomJoinedNotification(int(room), displayName)
		}

		return notifications.NewRoomJoinedNotificationWithName(int(room), displayName, name)
	}()

	return targets, notification, nil
}

func onNewFollower(event *pubsub.Event) ([]int, *notifications.PushNotification, error) {
	creator, err := getId(event, "follower")
	if err != nil {
		return nil, nil, err
	}

	targetID, ok := event.Params["id"].(float64)
	if !ok {
		return nil, nil, errors.New("failed to recover target ID")
	}

	displayName, err := getDisplayName(creator)
	if err != nil {
		return nil, nil, err
	}

	return []int{int(targetID)}, notifications.NewFollowerNotification(creator, displayName), nil
}

func onRoomInvite(event *pubsub.Event) ([]int, *notifications.PushNotification, error) {
	creator, err := getId(event, "from")
	if err != nil {
		return nil, nil, err
	}

	targetID, ok := event.Params["id"].(float64)
	if !ok {
		return nil, nil, errors.New("failed to recover target ID")
	}

	name := event.Params["name"].(string)
	room, ok := event.Params["room"].(float64)
	if !ok {
		return nil, nil, errors.New("failed to recover room ID")
	}

	displayName, err := getDisplayName(creator)
	if err != nil {
		return nil, nil, err
	}

	notification := func() *notifications.PushNotification {
		if name == "" {
			return notifications.NewRoomInviteNotification(int(room), displayName)
		}

		return notifications.NewRoomInviteNotificationWithName(int(room), displayName, name)
	}()

	return []int{int(targetID)}, notification, nil
}

func getId(event *pubsub.Event, field string) (int, error) {
	creator, ok := event.Params[field].(float64)
	if !ok {
		return 0, errors.New("failed to recover creator")
	}

	return int(creator), nil
}

func getCreatorId(event *pubsub.Event) (int, error) {
	return getId(event, "creator")
}

func getDisplayName(id int) (string, error) {
	user, err := userBackend.FindByID(id)
	if err != nil {
		return "", err
	}

	return user.DisplayName, nil
}
