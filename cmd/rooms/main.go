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

	"github.com/soapboxsocial/soapbox/pkg/api/middleware"
	"github.com/soapboxsocial/soapbox/pkg/blocks"
	"github.com/soapboxsocial/soapbox/pkg/groups"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
	roomGRPC "github.com/soapboxsocial/soapbox/pkg/rooms/grpc"
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
		Router: sfu.RouterConfig{
			AudioLevelFilter: 20,
			AudioLevelThreshold: 40,
			AudioLevelInterval: 1000,
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

	repository := rooms.NewRepository()
	sm := sessions.NewSessionManager(rdb)

	addr := "127.0.0.1:50052"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("failed to listen: %v", err)
		return
	}

	s := grpc.NewServer()
	pb.RegisterRoomServiceServer(
		s,
		roomGRPC.NewService(repository),
	)

	go func() {
		err = s.Serve(lis)
		if err != nil {
			log.Panicf("failed to serve: %v", err)
		}
	}()

	server := rooms.NewServer(
		sfu.NewSFU(config),
		sm,
		users.NewUserBackend(db),
		pubsub.NewQueue(rdb),
		rooms.NewCurrentRoomBackend(rdb),
		groups.NewBackend(db),
		repository,
		blocks.NewBackend(db),
	)

	endpoint := rooms.NewEndpoint(repository, server)
	router := endpoint.Router()

	amw := middleware.NewAuthenticationMiddleware(sm)
	router.Use(amw.Middleware)

	err = http.ListenAndServe(":8082", httputil.CORS(router))
	if err != nil {
		log.Print(err)
	}
}
