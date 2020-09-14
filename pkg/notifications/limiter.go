package notifications

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var keyNotificationCooldown = 30 * time.Minute
var valueString = "placeholder"

type Limiter struct {
	rdb *redis.Client
}

func NewLimiter(rdb *redis.Client) *Limiter {
	return &Limiter{
		rdb: rdb,
	}
}

func (l *Limiter) ShouldSendNotification(target string, args map[string]interface{}, category NotificationCategory) bool {
	if category == NEW_FOLLOWER {
		return true
	}

	res, err := l.rdb.Get(l.rdb.Context(), limiterKeyForRoom(target, args["id"].(int))).Result()
	if err != nil && res != valueString {
		return true
	}

	return false
}

func (l *Limiter) SentNotification(target string, args map[string]interface{}, category NotificationCategory) {
	if category == NEW_FOLLOWER {
		return
	}

	l.rdb.Set(l.rdb.Context(), limiterKeyForRoom(target, args["id"].(int)), valueString, keyNotificationCooldown)
}

func limiterKeyForRoom(target string, id int) string {
	return fmt.Sprintf("%s_room_%d", target, id)

}