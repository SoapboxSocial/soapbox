package main

import (
	"database/sql"
	"log"
	"net"
	"net/http"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	iLog "github.com/pion/ion-log"
	"github.com/pion/ion-sfu/pkg/sfu"
	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/groups"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

func main() {
	config := sfu.Config{
		WebRTC: sfu.WebRTCConfig{
			ICEServers: []sfu.ICEServerConfig{
				{
					URLs: []string{
						"stun:stun.l.google.com:19302",
						"stun:stun1.l.google.com:19302",
						"stun:stun2.l.google.com:19302",
						"stun:stun3.l.google.com:19302",
						"stun:stun4.l.google.com:19302",
					},
				},
			},
		},
		Log: iLog.Config{
			Level: "debug",
		},
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=voicely password=voicely dbname=voicely sslmode=disable")
	if err != nil {
		panic(err)
	}

	addr := ":50051"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("failed to listen: %v", err)
		return
	}

	repository := rooms.NewRepository()

	s := grpc.NewServer()
	pb.RegisterSFUServer(
		s,
		rooms.NewServer(
			sfu.NewSFU(config),
			sessions.NewSessionManager(rdb),
			users.NewUserBackend(db),
			pubsub.NewQueue(rdb),
			rooms.NewCurrentRoomBackend(rdb),
			groups.NewBackend(db),
			repository,
		),
	)

	endpoint := rooms.NewEndpoint(repository)
	router := endpoint.Router()

	go func() {
		err := http.ListenAndServe(":8082", httputil.CORS(router))
		if err != nil {
			log.Print(err)
		}
	}()

	err = s.Serve(lis)
	if err != nil {
		log.Panicf("failed to serve: %v", err)
	}
}
