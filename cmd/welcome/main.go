package main

// This command is responsible for registering welcome rooms when a new user signs up.

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	queue := pubsub.NewQueue(rdb)
	events := queue.Subscribe(pubsub.UserTopic)

	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	client := pb.NewRoomServiceClient(conn)

	for evt := range events {
		if evt.Type != pubsub.EventTypeNewUser {
			continue
		}

		id, err := evt.GetInt("id")
		if err != nil {
			log.Printf("evt.GetInt err %v", err)
			continue
		}

		resp, err := client.RegisterWelcomeRoom(context.Background(), &pb.WelcomeRoomRegisterRequest{UserId: int64(id)})
		if err != nil {
			log.Printf("client.RegisterWelcomeRoom err %v", err)
		}

		err = queue.Publish(pubsub.RoomTopic, pubsub.NewWelcomeRoomEvent(id, resp.Id))
		if err != nil {
			log.Printf("queue.Publish err: %v", err)
		}
	}
}
