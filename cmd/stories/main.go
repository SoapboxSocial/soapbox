package main

import (
	"flag"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/sql"
	"github.com/soapboxsocial/soapbox/pkg/stories"
)

type Conf struct {
	Data struct {
		Path string `mapstructure:"path"`
	} `mapstructure:"data"`
	DB conf.PostgresConf `mapstructure:"db"`
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

	db, err := sql.Open(config.DB)
	if err != nil {
		log.Fatalf("failed to open db: %s", err)
	}

	backend := stories.NewBackend(db)
	files := stories.NewFileBackend(config.Data.Path)

	now := time.Now().Unix()

	ids, err := backend.DeleteExpired(now)
	if err != nil {
		panic(err)
	}

	for _, id := range ids {
		err := files.Remove(id + ".aac")
		if err != nil {
			log.Printf("files.Remove err: %v\n", err)
		}
	}
}
