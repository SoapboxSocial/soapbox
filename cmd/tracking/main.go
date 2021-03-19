package main

import (
	"flag"
	"log"

	"github.com/dukex/mixpanel"

	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/redis"
	"github.com/soapboxsocial/soapbox/pkg/tracking/trackers"
)

type Conf struct {
	Mixpanel struct {
		Token string `mapstructure:"token"`
		URL   string `mapstructure:"url"`
	} `mapstructure:"mixpanel"`
	Redis conf.RedisConf `mapstructure:"redis"`
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

	client := mixpanel.New(config.Mixpanel.Token, config.Mixpanel.URL)
	tracker := trackers.NewMixpanelTracker(client)

	rdb := redis.NewRedis(config.Redis)
	queue := pubsub.NewQueue(rdb)

	events := queue.Subscribe(pubsub.RoomTopic, pubsub.UserTopic, pubsub.GroupTopic, pubsub.StoryTopic)

	for evt := range events {
		if !tracker.CanTrack(evt) {
			continue
		}

		err := tracker.Track(evt)
		if err != nil {
			log.Printf("tacker.Track err %v", err)
		}
	}
}
