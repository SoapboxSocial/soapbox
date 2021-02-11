package main

import (
	"errors"
	"flag"
	"log"
	"time"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"

	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/followers"
	"github.com/soapboxsocial/soapbox/pkg/groups"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/notifications/limiter"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/redis"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/sql"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

const TEST_ACCOUNT_ID = 19

var (
	errRoomPrivate = errors.New("room is private")
)

var (
	devicesBackend      *devices.Backend
	userBackend         *users.UserBackend
	followersBackend    *followers.FollowersBackend
	groupsBackend       *groups.Backend
	service             *notifications.Service
	notificationLimiter *limiter.Limiter
	notificationStorage *notifications.Storage
)

type handlerFunc func(*pubsub.Event) ([]int, *notifications.PushNotification, error)

type Conf struct {
	APNS  conf.AppleConf    `mapstructure:"apns"`
	Redis conf.RedisConf    `mapstructure:"redis"`
	DB    conf.PostgresConf `mapstructure:"db"`
}

func parse() (*Conf, error) {
	var file string
	flag.StringVar(&file, "c", "config.toml", "config file")

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

	rdb := redis.NewRedis(config.Redis)
	queue := pubsub.NewQueue(rdb)

	db, err := sql.Open(config.DB)
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}

	devicesBackend = devices.NewBackend(db)
	userBackend = users.NewUserBackend(db)
	followersBackend = followers.NewFollowersBackend(db)
	currentRoom := rooms.NewCurrentRoomBackend(rdb)
	notificationLimiter = limiter.NewLimiter(rdb, currentRoom)
	notificationStorage = notifications.NewStorage(rdb)
	groupsBackend = groups.NewBackend(db)

	authKey, err := token.AuthKeyFromFile(config.APNS.Path)
	if err != nil {
		panic(err)
	}

	// @todo add flag for which enviroment

	client := apns2.NewTokenClient(&token.Token{
		AuthKey: authKey,
		KeyID:   config.APNS.KeyID,
		TeamID:  config.APNS.TeamID,
	}).Production()

	service = notifications.NewService(config.APNS.Bundle, client)

	events := queue.Subscribe(pubsub.RoomTopic, pubsub.UserTopic, pubsub.GroupTopic)

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
		pushNotification(target, event, notification)
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
	case pubsub.EventTypeGroupInvite:
		return onGroupInvite
	case pubsub.EventTypeNewGroupRoom:
		return onGroupRoomCreation
	case pubsub.EventTypeWelcomeRoom:
		return onWelcomeRoom
	default:
		return nil
	}
}

func pushNotification(target int, event *pubsub.Event, notification *notifications.PushNotification) {
	if !notificationLimiter.ShouldSendNotification(target, event) {
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

	notificationLimiter.SentNotification(target, event)

	store := getNotificationForStore(notification)
	if store == nil {
		return
	}

	err = notificationStorage.Store(target, store)
	if err != nil {
		log.Printf("notificationStorage.Store err: %v\n", err)
	}
}

func getNotificationForStore(notification *notifications.PushNotification) *notifications.Notification {
	switch notification.Category {
	case notifications.NEW_FOLLOWER:
		return &notifications.Notification{
			Timestamp: time.Now().Unix(),
			From:      notification.Arguments["id"].(int),
			Category:  notification.Category,
		}
	case notifications.GROUP_INVITE:
		return &notifications.Notification{
			Timestamp: time.Now().Unix(),
			From:      notification.Arguments["from"].(int),
			Category:  notification.Category,
			Arguments: map[string]interface{}{"group": notification.Arguments["id"].(int)},
		}
	case notifications.WELCOME_ROOM:
		return &notifications.Notification{
			Timestamp: time.Now().Unix(),
			From:      notification.Arguments["from"].(int),
			Category:  notification.Category,
			Arguments: map[string]interface{}{"room": notification.Arguments["id"]},
		}
	default:
		return nil
	}
}

func onRoomCreation(event *pubsub.Event) ([]int, *notifications.PushNotification, error) {
	if pubsub.RoomVisibility(event.Params["visibility"].(string)) == pubsub.Private {
		return nil, nil, errRoomPrivate
	}

	creator, err := event.GetInt("creator")
	if err != nil {
		return nil, nil, err
	}

	if creator == TEST_ACCOUNT_ID {
		return nil, nil, nil
	}

	targets, err := followersBackend.GetAllFollowerIDsFor(creator)
	if err != nil {
		return nil, nil, err
	}

	name := event.Params["name"].(string)
	room := event.Params["id"].(string)

	displayName, err := getDisplayName(creator)
	if err != nil {
		return nil, nil, err
	}

	notification := func() *notifications.PushNotification {
		if name == "" {
			return notifications.NewRoomNotification(room, displayName)
		}

		return notifications.NewRoomNotificationWithName(room, displayName, name)
	}()

	return targets, notification, nil
}
func onGroupRoomCreation(event *pubsub.Event) ([]int, *notifications.PushNotification, error) {
	creator, err := event.GetInt("creator")
	if err != nil {
		return nil, nil, err
	}

	if creator == TEST_ACCOUNT_ID {
		return nil, nil, nil
	}

	groupId, err := event.GetInt("group")
	if err != nil {
		return nil, nil, err
	}

	group, err := groupsBackend.FindById(groupId)
	if err != nil {
		return nil, nil, err
	}

	targets := make([]int, 0)

	if group.GroupType == "public" {
		followerIDs, err := followersBackend.GetAllFollowerIDsFor(creator)
		if err != nil {
			return nil, nil, err
		}

		targets = append(targets, followerIDs...)
	}

	memberIDs, err := groupsBackend.GetAllMemberIds(groupId, creator)
	if err != nil {
		return nil, nil, err
	}

	targets = append(targets, memberIDs...)

	name := event.Params["name"].(string)
	room := event.Params["id"].(string)

	displayName, err := getDisplayName(creator)
	if err != nil {
		return nil, nil, err
	}

	notification := func() *notifications.PushNotification {
		if name == "" {
			return notifications.NewRoomWithGroupNotification(room, displayName, group.Name)
		}

		return notifications.NewRoomWithGroupAndNameNotification(room, displayName, group.Name, name)
	}()

	return targets, notification, nil
}

func onRoomJoined(event *pubsub.Event) ([]int, *notifications.PushNotification, error) {
	if pubsub.RoomVisibility(event.Params["visibility"].(string)) == pubsub.Private {
		return nil, nil, errRoomPrivate
	}

	creator, err := event.GetInt("creator")
	if err != nil {
		return nil, nil, err
	}

	targets, err := followersBackend.GetAllFollowerIDsFor(creator)
	if err != nil {
		return nil, nil, err
	}

	name := event.Params["name"].(string)
	room := event.Params["id"].(string)

	displayName, err := getDisplayName(creator)
	if err != nil {
		return nil, nil, err
	}

	notification := func() *notifications.PushNotification {
		if name == "" {
			return notifications.NewRoomJoinedNotification(room, displayName)
		}

		return notifications.NewRoomJoinedNotificationWithName(room, displayName, name)
	}()

	return targets, notification, nil
}

func onNewFollower(event *pubsub.Event) ([]int, *notifications.PushNotification, error) {
	creator, err := event.GetInt("follower")
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

func onGroupInvite(event *pubsub.Event) ([]int, *notifications.PushNotification, error) {
	creator, err := event.GetInt("from")
	if err != nil {
		return nil, nil, err
	}

	targetID, err := event.GetInt("id")
	if err != nil {
		return nil, nil, err
	}

	groupId, err := event.GetInt("group")
	if err != nil {
		return nil, nil, err
	}

	displayName, err := getDisplayName(creator)
	if err != nil {
		return nil, nil, err
	}

	group, err := getGroupName(groupId)
	if err != nil {
		return nil, nil, err
	}

	return []int{targetID}, notifications.NewGroupInviteNotification(groupId, creator, displayName, group), nil
}

func onRoomInvite(event *pubsub.Event) ([]int, *notifications.PushNotification, error) {
	creator, err := event.GetInt("from")
	if err != nil {
		return nil, nil, err
	}

	targetID, ok := event.Params["id"].(float64)
	if !ok {
		return nil, nil, errors.New("failed to recover target ID")
	}

	name := event.Params["name"].(string)
	room := event.Params["room"].(string)

	displayName, err := getDisplayName(creator)
	if err != nil {
		return nil, nil, err
	}

	notification := func() *notifications.PushNotification {
		if name == "" {
			return notifications.NewRoomInviteNotification(room, displayName)
		}

		return notifications.NewRoomInviteNotificationWithName(room, displayName, name)
	}()

	return []int{int(targetID)}, notification, nil
}

func onWelcomeRoom(event *pubsub.Event) ([]int, *notifications.PushNotification, error) {
	creator, err := event.GetInt("id")
	if err != nil {
		return nil, nil, err
	}

	room := event.Params["room"].(string)

	displayName, err := getDisplayName(creator)
	if err != nil {
		return nil, nil, err
	}

	staticTargets := []int{1, 75, 962, 13, 6}

	notification := notifications.NewWelcomeRoomNotification(displayName, room, creator)
	return staticTargets, notification, nil
}

func getDisplayName(id int) (string, error) {
	user, err := userBackend.FindByID(id)
	if err != nil {
		return "", err
	}

	return user.DisplayName, nil
}

func getGroupName(id int) (string, error) {
	group, err := groupsBackend.FindById(id)
	if err != nil {
		return "", err
	}

	return group.Name, nil
}
