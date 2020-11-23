package limiter

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
)

var (
	followerCooldown   = 30 * time.Minute
	roomInviteCooldown = 5 * time.Minute
	roomMemberCooldown = 5 * time.Minute
	roomCooldown       = 30 * time.Minute
)

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
	switch event.Type {
	case pubsub.EventTypeNewRoom, pubsub.EventTypeRoomJoin:
		// 30 minutes for any notification with the same room ID
		if l.isLimited(limiterKeyForRoom(target, event)) {
			return false
		}

		// 5 minutes for any notification with the same room member
		if l.isLimited(limiterKeyForRoomMember(target, event)) {
			return false
		}
	case pubsub.EventTypeRoomInvite:
		return !l.isLimited(limiterKeyForRoomInvite(target, event))
	case pubsub.EventTypeNewFollower:
		return !l.isLimited(limiterKeyForFollowerEvent(target, event))
	case pubsub.EventTypeGroupInvite:
		return true
	default:
		return false
	}

	return false
}

func (l *Limiter) SentNotification(target int, event *pubsub.Event) {
	switch event.Type {
	case pubsub.EventTypeNewRoom, pubsub.EventTypeRoomJoin:
		l.limit(limiterKeyForRoomMember(target, event), roomMemberCooldown)
		l.limit(limiterKeyForRoom(target, event), roomCooldown)
	case pubsub.EventTypeRoomInvite:
		l.limit(limiterKeyForFollowerEvent(target, event), roomInviteCooldown)
	case pubsub.EventTypeNewFollower:
		l.limit(limiterKeyForFollowerEvent(target, event), followerCooldown)
	}
}

const placeholder = "placeholder"

func (l *Limiter) isLimited(key string) bool {
	res, _ := l.rdb.Get(l.rdb.Context(), key).Result()
	return res == placeholder
}

func (l *Limiter) limit(key string, duration time.Duration) {
	l.rdb.Set(l.rdb.Context(), key, placeholder, duration)
}

func limiterKeyForRoom(target int, event *pubsub.Event) string {
	return fmt.Sprintf("notifications_%d_room_%d", target, event.Params["id"])
}

func limiterKeyForRoomMember(target int, event *pubsub.Event) string {
	return fmt.Sprintf("notifications_%d_room_member_%d", target, event.Params["creator"])
}

func limiterKeyForRoomInvite(target int, event *pubsub.Event) string {
	return fmt.Sprintf("notifications_%d_room_invite_%d", target, event.Params["room"])
}

func limiterKeyForFollowerEvent(target int, event *pubsub.Event) string {
	return fmt.Sprintf("notifications_%d_follower_%d", target, event.Params["follower"])
}
