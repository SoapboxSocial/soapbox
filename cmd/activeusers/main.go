package main

import (
	"flag"
	"log"

	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/redis"
)

func parse() (*conf.RedisConf, error) {
	var file string
	flag.StringVar(&file, "c", "config.toml", "config file")
	flag.Parse()

	config := &conf.RedisConf{}
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
	rdb := redis.NewRedis(*config)
	queue := pubsub.NewQueue(rdb)

	events := queue.Subscribe(pubsub.UserTopic)

	for evt := range events {
		if evt.Type != pubsub.EventTypeUserHeartbeat {
			continue
		}

		// @TODO write
	}
}
