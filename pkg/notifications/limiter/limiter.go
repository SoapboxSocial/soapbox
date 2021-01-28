package limiter

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
)

var (
	followerCooldown   = 15 * time.Minute
	roomInviteCooldown = 5 * time.Minute
	roomMemberCooldown = 5 * time.Minute
	roomCooldown       = 10 * time.Minute
	groupRoomCooldown  = 10 * time.Minute
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
	case pubsub.EventTypeNewGroupRoom:
		if l.isLimited(limiterKeyForGroupRoom(target, event)) {
			return false
		}

		fallthrough
	case pubsub.EventTypeNewRoom, pubsub.EventTypeRoomJoin:
		// 30 minutes for any notification with the same room ID
		if l.isLimited(limiterKeyForRoom(target, event)) {
			return false
		}

		// 5 minutes for any notification with the same room member
		if l.isLimited(limiterKeyForRoomMember(target, event)) {
			return false
		}

		if l.isUserInRoom(target, event) {
			return false
		}

		return true
	case pubsub.EventTypeRoomInvite:
		return !l.isLimited(limiterKeyForRoomInvite(target, event))
	case pubsub.EventTypeNewFollower:
		return !l.isLimited(limiterKeyForFollowerEvent(target, event))
	case pubsub.EventTypeGroupInvite:
		return true
	default:
		return false
	}
}

func (l *Limiter) SentNotification(target int, event *pubsub.Event) {
	switch event.Type {
	case pubsub.EventTypeNewGroupRoom:
		l.limit(limiterKeyForGroupRoom(target, event), groupRoomCooldown)
		fallthrough
	case pubsub.EventTypeNewRoom, pubsub.EventTypeRoomJoin:
		l.limit(limiterKeyForRoomMember(target, event), roomMemberCooldown)
		l.limit(limiterKeyForRoom(target, event), roomCooldown)
	case pubsub.EventTypeRoomInvite:
		l.limit(limiterKeyForRoomInvite(target, event), roomInviteCooldown)
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

func (l *Limiter) isUserInRoom(user int, event *pubsub.Event) bool {
	room, _ := l.currentRoom.GetCurrentRoomForUser(user)
	if room == "" {
		return false
	}

	id := event.Params["id"].(string)

	return id == room
}

func limiterKeyForGroupRoom(target int, event *pubsub.Event) string {
	return fmt.Sprintf("notifications_limit_%d_room_in_group_%v", target, event.Params["group"])
}

func limiterKeyForRoom(target int, event *pubsub.Event) string {
	return fmt.Sprintf("notifications_limit_%d_room_%v", target, event.Params["id"])
}

func limiterKeyForRoomMember(target int, event *pubsub.Event) string {
	return fmt.Sprintf("notifications_limit_%d_room_member_%v", target, event.Params["creator"])
}

func limiterKeyForRoomInvite(target int, event *pubsub.Event) string {
	return fmt.Sprintf("notifications_limit_%d_room_invite_%v", target, event.Params["room"])
}

func limiterKeyForFollowerEvent(target int, event *pubsub.Event) string {
	return fmt.Sprintf("notifications_limit_%d_follower_%v", target, event.Params["follower"])
}

func getInt(event *pubsub.Event, value string) (int, error) {
	id, ok := event.Params[value].(float64)
	if !ok {
		return 0, fmt.Errorf("failed to recover %s", value)
	}

	return int(id), nil
}
