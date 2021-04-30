package cmd

import (
	sqldb "database/sql"
	"fmt"
	"log"
	"net"

	"github.com/pkg/errors"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/token"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/apple"
	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	notificationsGRPC "github.com/soapboxsocial/soapbox/pkg/notifications/grpc"
	"github.com/soapboxsocial/soapbox/pkg/notifications/handlers"
	"github.com/soapboxsocial/soapbox/pkg/notifications/pb"
	"github.com/soapboxsocial/soapbox/pkg/notifications/worker"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/redis"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
	roompb "github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/sql"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Conf struct {
	APNS  conf.AppleConf    `mapstructure:"apns"`
	Redis conf.RedisConf    `mapstructure:"redis"`
	DB    conf.PostgresConf `mapstructure:"db"`
	Rooms conf.AddrConf     `mapstructure:"rooms"`
	GRPC  conf.AddrConf     `mapstructure:"GRPC"`
}

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "runs a notification worker",
	RunE:  runWorker,
}

var file string

func init() {
	workerCmd.Flags().StringVarP(&file, "config", "c", "config.toml", "config file")
}

func runWorker(*cobra.Command, []string) error {
	config := &Conf{}
	err := conf.Load(file, config)
	if err != nil {
		return errors.Wrap(err, "failed to parse config")
	}

	rdb := redis.NewRedis(config.Redis)
	queue := pubsub.NewQueue(rdb)

	db, err := sql.Open(config.DB)
	if err != nil {
		return errors.Wrap(err, "failed to open db")
	}

	currentRoom := rooms.NewCurrentRoomBackend(db)

	authKey, err := token.AuthKeyFromFile(config.APNS.Path)
	if err != nil {
		return errors.Wrap(err, "failed to load auth key")
	}

	// @todo add flag for which enviroment

	client := apns2.NewTokenClient(&token.Token{
		AuthKey: authKey,
		KeyID:   config.APNS.KeyID,
		TeamID:  config.APNS.TeamID,
	}).Production()

	settings := notifications.NewSettings(db)
	notificationHandlers := setupHandlers(db, config.Rooms, settings)

	events := queue.Subscribe(pubsub.RoomTopic, pubsub.UserTopic)

	dispatch := worker.NewDispatcher(5, &worker.Config{
		APNS:    apple.NewAPNS(config.APNS.Bundle, client),
		Limiter: notifications.NewLimiter(rdb, currentRoom),
		Devices: devices.NewBackend(db),
		Store:   notifications.NewStorage(rdb),
	})
	dispatch.Run()

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
				dispatch.Dispatch(target, notification)
			}
		}(event)
	}

	return runServer(config.GRPC, dispatch, settings)
}

func runServer(addr conf.AddrConf, dispatcher *worker.Dispatcher, settings *notifications.Settings) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", addr.Host, addr.Port))
	if err != nil {
		return errors.Wrap(err, "failed to start server")
	}

	gs := grpc.NewServer()
	pb.RegisterNotificationServiceServer(
		gs,
		notificationsGRPC.NewService(dispatcher, settings),
	)

	err = gs.Serve(lis)
	if err != nil {
		log.Panicf("failed to serve: %v", err)
	}

	return nil
}

func setupHandlers(db *sqldb.DB, roomsAddr conf.AddrConf, settings *notifications.Settings) map[pubsub.EventType]handlers.Handler {
	userBackend := users.NewUserBackend(db)

	notificationHandlers := make(map[pubsub.EventType]handlers.Handler)

	followers := handlers.NewFollowerNotificationHandler(settings, userBackend)
	notificationHandlers[followers.Type()] = followers

	creation := handlers.NewRoomCreationNotificationHandler(settings, userBackend)
	notificationHandlers[creation.Type()] = creation

	invite := handlers.NewRoomInviteNotificationHandler(settings, userBackend)
	notificationHandlers[invite.Type()] = invite

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", roomsAddr.Host, roomsAddr.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	join := handlers.NewRoomJoinNotificationHandler(settings, roompb.NewRoomServiceClient(conn))
	notificationHandlers[join.Type()] = join

	welcome := handlers.NewWelcomeRoomNotificationHandler(userBackend)
	notificationHandlers[welcome.Type()] = welcome

	return notificationHandlers
}
