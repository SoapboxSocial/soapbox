package notifications

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var keyNotificationCooldown = 30 * time.Minute
var roomInviteNotificationCooldown = 5 * time.Minute
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
	if category == NEW_FOLLOWER || category == NEW_ROOM {
		return true
	}

	if category == ROOM_JOINED {
		return !l.isLimited(limiterKeyForRoom(target, args["id"].(int)))
	}

	if category == ROOM_INVITE {
		return !l.isLimited(limiterKeyForRoomInvite(target, args["id"].(int)))
	}

	return false
}

func (l *Limiter) SentNotification(target string, args map[string]interface{}, category NotificationCategory) {
	if category == NEW_FOLLOWER {
		return
	}

	if category == ROOM_JOINED || category == NEW_ROOM {
		l.rdb.Set(l.rdb.Context(), limiterKeyForRoom(target, args["id"].(int)), valueString, keyNotificationCooldown)
		return
	}

	if category == ROOM_INVITE {
		l.rdb.Set(l.rdb.Context(), limiterKeyForRoomInvite(target, args["id"].(int)), valueString, roomInviteNotificationCooldown)
		return
	}



}

func (l *Limiter) isLimited(key string) bool {
	res, err := l.rdb.Get(l.rdb.Context(), key).Result()
	if err != nil || res != valueString {
		return false
	}

	return true
}

func limiterKeyForRoom(target string, id int) string {
	return fmt.Sprintf("%s_room_%d", target, id)
}

func limiterKeyForRoomInvite(target string, id int) string {
	return fmt.Sprintf("%s_room_invite_%d", target, id)
}
