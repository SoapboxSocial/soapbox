package main

// This command is responsible for registering welcome rooms when a new user signs up.

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/redis"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

var queue *pubsub.Queue
var client pb.RoomServiceClient

type Conf struct {
	Redis conf.RedisConf `mapstructure:"redis"`
	Rooms conf.AddrConf  `mapstructure:"rooms"`
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

	queue = pubsub.NewQueue(rdb)
	events := queue.Subscribe(pubsub.UserTopic)

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", config.Rooms.Host, config.Rooms.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	client = pb.NewRoomServiceClient(conn)

	for evt := range events {
		if evt.Type != pubsub.EventTypeNewUser {
			continue
		}

		go sendNotification(evt)
	}
}

func sendNotification(event *pubsub.Event) {
	time.Sleep(1 * time.Minute)

	id, err := event.GetInt("id")
	if err != nil {
		log.Printf("evt.GetInt err %v", err)
		return
	}

	resp, err := client.RegisterWelcomeRoom(context.Background(), &pb.WelcomeRoomRegisterRequest{UserId: int64(id)})
	if err != nil {
		log.Printf("client.RegisterWelcomeRoom err %v", err)
		return
	}

	err = queue.Publish(pubsub.RoomTopic, pubsub.NewWelcomeRoomEvent(id, resp.Id))
	if err != nil {
		log.Printf("queue.Publish err: %v", err)
	}
}
