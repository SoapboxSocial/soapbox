package limiter

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
)

var keyNotificationCooldown = 30 * time.Minute
var valueString = "placeholder"

type Limiter struct {
	rdb         *redis.Client
	currentRoom rooms.CurrentRoomBackend
}

func NewLimiter(rdb *redis.Client) *Limiter {
	return &Limiter{
		rdb: rdb,
	}
}

func (l *Limiter) ShouldSendNotification(target devices.Device, args map[string]interface{}, category notifications.NotificationCategory) bool {
	if category == notifications.NEW_FOLLOWER || category == notifications.NEW_ROOM {
		return true
	}

	id := args["id"].(int)
	room, _ := l.currentRoom.GetCurrentRoomForUser(target.ID)
	if room == id {
		return false
	}

	res, err := l.rdb.Get(l.rdb.Context(), limiterKeyForRoom(target.ID, id)).Result()
	if err != nil || res != valueString {
		return true
	}

	return false
}

func (l *Limiter) SentNotification(target devices.Device, args map[string]interface{}, category notifications.NotificationCategory) {
	if category == notifications.NEW_FOLLOWER {
		return
	}

	l.rdb.Set(l.rdb.Context(), limiterKeyForRoom(target.ID, args["id"].(int)), valueString, keyNotificationCooldown)
}

func limiterKeyForRoom(target, id int) string {
	return fmt.Sprintf("%d_room_%d", target, id)
}
