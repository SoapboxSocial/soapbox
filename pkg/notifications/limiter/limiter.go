package limiter

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
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
		rdb:         rdb,
		currentRoom: currentRoom,
	}
}

func (l *Limiter) ShouldSendNotification(target int, event *pubsub.Event) bool {
	// @TODO WE SHOULD LIMIT NEW FOLLOWER NOTIFICATION SO SOMEONE CANT REFOLLOW AND SPAM
	//if category == notifications.NEW_FOLLOWER || category == notifications.GROUP_INVITE {
	//	return true
	//}
	//
	//if category == notifications.NEW_ROOM {
	//	return !l.isLimited(limiterKeyForRoom(target, args["id"].(int)))
	//}
	//
	//id := args["id"].(int)
	//room, _ := l.currentRoom.GetCurrentRoomForUser(target)
	//if room == id {
	//	return false
	//}
	//
	//if category == notifications.ROOM_JOINED {
	//	return !l.isLimited(limiterKeyForRoom(target, args["id"].(int)))
	//}
	//
	//if category == notifications.ROOM_INVITE {
	//	return !l.isLimited(limiterKeyForRoomInvite(target, args["id"].(int)))
	//}

	return false
}

func (l *Limiter) SentNotification(target int, event *pubsub.Event) {

	// @TODO CHECK IF KEYS ARE OK?

	//if event.Type == pubsub.EventTypeRoomJoin || event.Type == pubsub.EventTypeNewRoom {
	//	l.rdb.Set(l.rdb.Context(), limiterKeyForRoom(target, event.Params["id"].(int)), valueString, keyNotificationCooldown)
	//	return
	//}
	//
	//if event.Type == pubsub.EventTypeRoomInvite {
	//	l.rdb.Set(l.rdb.Context(), limiterKeyForRoomInvite(target, event.Params["id"].(int)), valueString, roomInviteNotificationCooldown)
	//}
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
