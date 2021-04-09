package redis_test

import (
	"strconv"
	"testing"

	"github.com/alicebob/miniredis"

	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/redis"
)

func TestNewRedis(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	port, err := strconv.Atoi(mr.Port())
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewRedis(conf.RedisConf{
		Port: port,
		Host: mr.Host(),
	})

	val, err := rdb.Ping(rdb.Context()).Result()
	if err != nil {
		t.Fatal(err)
	}

	if val != "PONG" {
		t.Fatalf("unexpected val %s", val)
	}
}
