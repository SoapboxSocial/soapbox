// Package redis contains helper functions for working with redis.
package redis

import (
	"crypto/tls"
	"fmt"

	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/conf"
)

// NewRedis returns a new redis instance created using the config
func NewRedis(config conf.RedisConf) *redis.Client {

	opts := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.Database,
	}

	if !config.DisableTLS {
		opts.TLSConfig = &tls.Config{}
	}

	return redis.NewClient(opts)
}
