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
var roomInviteNotificationCooldown = 5 * time.Minute

var valueString = "placeholder"

type Limiter struct {
	rdb         *redis.Client
	currentRoom *rooms.CurrentRoomBackend
}

func NewLimiter(rdb *redis.Client, currentRoom *rooms.CurrentRoomBackend) *Limiter {
	return &Limiter{
		rdb: rdb,
		currentRoom: currentRoom,
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

	if category == notifications.ROOM_JOINED {
		return !l.isLimited(limiterKeyForRoom(target.ID, args["id"].(int)))
	}

	if category == notifications.ROOM_INVITE {
		return !l.isLimited(limiterKeyForRoomInvite(target.ID, args["id"].(int)))
	}

	return false
}

func (l *Limiter) SentNotification(target devices.Device, args map[string]interface{}, category notifications.NotificationCategory) {
	if category == notifications.NEW_FOLLOWER {
		return
	}

	if category == notifications.ROOM_JOINED || category == notifications.NEW_ROOM {
		l.rdb.Set(l.rdb.Context(), limiterKeyForRoom(target.ID, args["id"].(int)), valueString, keyNotificationCooldown)
		return
	}

	if category == notifications.ROOM_INVITE {
		l.rdb.Set(l.rdb.Context(), limiterKeyForRoomInvite(target.ID, args["id"].(int)), valueString, roomInviteNotificationCooldown)
		return
	}
}

func (l *Limiter) isLimited(key string) bool {
	res, _ := l.rdb.Get(l.rdb.Context(), key).Result()
	return res == valueString
}

func limiterKeyForRoom(target, id int) string {
	return fmt.Sprintf("%d_room_%d", target, id)
}

func limiterKeyForRoomInvite(target, id int) string {
	return fmt.Sprintf("%d_room_invite_%d", target, id)
}
