package main

import (
	sqldb "database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"
	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/apple"
	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/notifications/handlers"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/redis"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/sql"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Conf struct {
	APNS  conf.AppleConf    `mapstructure:"apns"`
	Redis conf.RedisConf    `mapstructure:"redis"`
	DB    conf.PostgresConf `mapstructure:"db"`
	GRPC  conf.AddrConf     `mapstructure:"grpc"`
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

	rdb := redis.NewRedis(config.Redis)
	queue := pubsub.NewQueue(rdb)

	db, err := sql.Open(config.DB)
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}

	currentRoom := rooms.NewCurrentRoomBackend(db)

	authKey, err := token.AuthKeyFromFile(config.APNS.Path)
	if err != nil {
		log.Fatal(err)
	}

	// @todo add flag for which enviroment

	client := apns2.NewTokenClient(&token.Token{
		AuthKey: authKey,
		KeyID:   config.APNS.KeyID,
		TeamID:  config.APNS.TeamID,
	}).Production()

	service := notifications.NewService(
		apple.NewAPNS(config.APNS.Bundle, client),
		notifications.NewLimiter(rdb, currentRoom),
		devices.NewBackend(db),
		notifications.NewStorage(rdb),
	)

	notificationHandlers := setupHandlers(db, config.GRPC)

	events := queue.Subscribe(pubsub.RoomTopic, pubsub.UserTopic)

	for event := range events {
		go func(event *pubsub.Event) {
			h := notificationHandlers[event.Type]
			if h == nil {
				return
			}

			targets, err := h.Targets(event)
			if err != nil {
				log.Printf("failed to get targets: %s", err)
				return
			}

			if len(targets) == 0 {
				log.Printf("no targets for: %d", event.Type)
				return
			}

			notification, err := h.Build(event)
			if err != nil {
				log.Printf("failed to build notifcation: %s", err)
				return
			}

			log.Printf("pushing %s to %d targets", notification.Category, len(targets))

			for _, target := range targets {
				service.Send(target, event, notification)
			}
		}(event)
	}
}

func setupHandlers(db *sqldb.DB, roomsAddr conf.AddrConf) map[pubsub.EventType]handlers.Handler {
	userBackend := users.NewUserBackend(db)
	targets := notifications.NewSettings(db)

	notificationHandlers := make(map[pubsub.EventType]handlers.Handler)

	followers := handlers.NewFollowerNotificationHandler(targets, userBackend)
	notificationHandlers[followers.Type()] = followers

	creation := handlers.NewRoomCreationNotificationHandler(targets, userBackend)
	notificationHandlers[creation.Type()] = creation

	invite := handlers.NewRoomInviteNotificationHandler(targets, userBackend)
	notificationHandlers[invite.Type()] = invite

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", roomsAddr.Host, roomsAddr.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	join := handlers.NewRoomJoinNotificationHandler(targets, pb.NewRoomServiceClient(conn))
	notificationHandlers[join.Type()] = join

	welcome := handlers.NewWelcomeRoomNotificationHandler(userBackend)
	notificationHandlers[welcome.Type()] = welcome

	return notificationHandlers
}
