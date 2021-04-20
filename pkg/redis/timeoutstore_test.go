package redis_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis"

	"github.com/soapboxsocial/soapbox/pkg/conf"
	"github.com/soapboxsocial/soapbox/pkg/redis"
)

func TestTimeoutStore(t *testing.T) {
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

	ts := redis.NewTimeoutStore(rdb)

	key := "foo"

	if ts.IsOnTimeout(key) {
		t.Fatal("should not be on timeout")
	}

	err = ts.SetTimeout(key, 5*time.Minute)
	if err != nil {
		t.Fatal(err)
	}

	if !ts.IsOnTimeout(key) {
		t.Fatal("key is on timeout")
	}
}
