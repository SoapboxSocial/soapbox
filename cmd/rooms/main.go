package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/pion/ion-sfu/pkg/middlewares/datachannel"
	"github.com/pion/ion-sfu/pkg/sfu"
	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/api/middleware"
	"github.com/soapboxsocial/soapbox/pkg/blocks"
	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/groups"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
	roomGRPC "github.com/soapboxsocial/soapbox/pkg/rooms/grpc"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/sql"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Conf struct {
	SFU   sfu.Config        `mapstructure:"sfu"`
	Redis conf.RedisConf    `mapstructure:"redis"`
	DB    conf.PostgresConf `mapstructure:"db"`
	GRPC  conf.AddrConf     `mapstructure:"grpc"`
	API   conf.AddrConf     `mapstructure:"api"`
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

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.Database,
	})

	db, err := sql.Open(config.DB)
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}

	repository := rooms.NewRepository()
	sm := sessions.NewSessionManager(rdb)
	ws := rooms.NewWelcomeStore(rdb)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.GRPC.Host, config.GRPC.Port))
	if err != nil {
		log.Panicf("failed to listen: %v", err)
		return
	}

	gs := grpc.NewServer()
	pb.RegisterRoomServiceServer(
		gs,
		roomGRPC.NewService(repository, ws),
	)

	go func() {
		err = gs.Serve(lis)
		if err != nil {
			log.Panicf("failed to serve: %v", err)
		}
	}()

	s := sfu.NewSFU(config.SFU)
	dc := s.NewDatachannel(sfu.APIChannelLabel)
	dc.Use(datachannel.SubscriberAPI)

	server := rooms.NewServer(
		s,
		sm,
		users.NewUserBackend(db),
		pubsub.NewQueue(rdb),
		rooms.NewCurrentRoomBackend(rdb),
		ws,
		groups.NewBackend(db),
		repository,
		blocks.NewBackend(db),
	)

	endpoint := rooms.NewEndpoint(repository, server)
	router := endpoint.Router()

	amw := middleware.NewAuthenticationMiddleware(sm)
	router.Use(amw.Middleware)

	err = http.ListenAndServe(fmt.Sprintf(":%d", config.API.Port), httputil.CORS(router))
	if err != nil {
		log.Print(err)
	}
}
