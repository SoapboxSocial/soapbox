package cmd

import (
	"fmt"
	"log"
	"net"
	"net/http"

	plog "github.com/pion/ion-sfu/pkg/logger"
	"github.com/pion/ion-sfu/pkg/middlewares/datachannel"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/blocks"
	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/groups"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/http/middlewares"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/redis"
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

var server = &cobra.Command{
	Use:   "server",
	Short: "runs a room server",
	RunE:  runServer,
}

var file string

func init() {
	server.Flags().StringVarP(&file, "config", "c", "config.toml", "config file")
}

func runServer(*cobra.Command, []string) error {
	config := &Conf{}
	err := conf.Load(file, config)
	if err != nil {
		return errors.Wrap(err, "failed to parse config")
	}

	rdb := redis.NewRedis(config.Redis)

	db, err := sql.Open(config.DB)
	if err != nil {
		return errors.Wrap(err, "failed to open db")
	}

	repository := rooms.NewRepository()
	sm := sessions.NewSessionManager(rdb)
	ws := rooms.NewWelcomeStore(rdb)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.GRPC.Host, config.GRPC.Port))
	if err != nil {
		return errors.Wrap(err, "failed to listen")
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

	plog.SetGlobalOptions(plog.GlobalConfig{V: 1})
	logger := plog.New()

	// SFU instance needs to be created with logr implementation
	sfu.Logger = logger

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

	amw := middlewares.NewAuthenticationMiddleware(sm)
	router.Use(amw.Middleware)

	return http.ListenAndServe(fmt.Sprintf(":%d", config.API.Port), httputil.CORS(router))
}
