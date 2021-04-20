package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/conf"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/metadata"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/sql"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Conf struct {
	DB     conf.PostgresConf `mapstructure:"db"`
	GRPC   conf.AddrConf     `mapstructure:"grpc"`
	Listen conf.AddrConf     `mapstructure:"listen"`
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

	db, err := sql.Open(config.DB)
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}

	usersBackend := users.NewUserBackend(db)

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", config.GRPC.Host, config.GRPC.Port), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	client := pb.NewRoomServiceClient(conn)
	endpoint := metadata.NewEndpoint(usersBackend, client)

	router := endpoint.Router()

	log.Print(http.ListenAndServe(fmt.Sprintf(":%d", config.Listen.Port), httputil.CORS(router)))
}
