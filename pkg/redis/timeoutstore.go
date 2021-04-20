package redis

import (
	"time"

	"github.com/go-redis/redis/v8"
)

const timeout = "timeout"

type TimeoutStore struct {
	rdb *redis.Client
}

func NewTimeoutStore(rdb *redis.Client) *TimeoutStore {
	return &TimeoutStore{rdb: rdb}
}

func (t *TimeoutStore) SetTimeout(key string, expiration time.Duration) error {
	return t.rdb.Set(t.rdb.Context(), key, timeout, expiration).Err()
}

func (t *TimeoutStore) IsOnTimeout(key string) bool {
	val, err := t.rdb.Get(t.rdb.Context(), key).Result()
	if err != nil {
		return false
	}

	return val == timeout
}
