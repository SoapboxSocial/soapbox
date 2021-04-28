package notifications

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
)

var (
	followerCooldown   = 15 * time.Minute
	roomInviteCooldown = 3 * time.Minute
	roomMemberCooldown = 5 * time.Minute
	roomCooldown       = 10 * time.Minute
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

func (l *Limiter) ShouldSendNotification(target Target, event *pubsub.Event) bool {
	switch event.Type {
	case pubsub.EventTypeNewRoom, pubsub.EventTypeRoomJoin:
		if target.RoomFrequency == FrequencyOff {
			return false
		}

		if l.isLimited(limiterKeyForRoom(target.ID, event)) {
			return false
		}

		if l.isLimited(limiterKeyForRoomMember(target.ID, event)) {
			return false
		}

		if l.isUserInRoom(target.ID, event) {
			return false
		}

		return true
	case pubsub.EventTypeRoomInvite:
		return !l.isLimited(limiterKeyForRoomInvite(target.ID, event))
	case pubsub.EventTypeNewFollower:
		if !target.Follows {
			return false
		}

		return !l.isLimited(limiterKeyForFollowerEvent(target.ID, event))
	case pubsub.EventTypeWelcomeRoom:
		return true // @TODO
	default:
		return false
	}
}

func (l *Limiter) SentNotification(target Target, event *pubsub.Event) {
	switch event.Type {
	case pubsub.EventTypeNewRoom:
		l.limit(limiterKeyForRoomMember(target.ID, event), getLimitForRoomFrequency(target.RoomFrequency, roomMemberCooldown))
	case pubsub.EventTypeRoomJoin:
		l.limit(limiterKeyForRoomMember(target.ID, event), getLimitForRoomFrequency(target.RoomFrequency, roomMemberCooldown))
		l.limit(limiterKeyForRoom(target.ID, event), getLimitForRoomFrequency(target.RoomFrequency, roomCooldown))
	case pubsub.EventTypeRoomInvite:
		l.limit(limiterKeyForRoomInvite(target.ID, event), roomInviteCooldown)
	case pubsub.EventTypeNewFollower:
		l.limit(limiterKeyForFollowerEvent(target.ID, event), followerCooldown)
	}
}

const limitPlaceholder = "placeholder"

func (l *Limiter) isLimited(key string) bool {
	res, _ := l.rdb.Get(l.rdb.Context(), key).Result()
	return res == limitPlaceholder
}

func (l *Limiter) limit(key string, duration time.Duration) {
	l.rdb.Set(l.rdb.Context(), key, limitPlaceholder, duration)
}

func (l *Limiter) isUserInRoom(user int, event *pubsub.Event) bool {
	room, _ := l.currentRoom.GetCurrentRoomForUser(user)
	if room == "" {
		return false
	}

	id := event.Params["id"].(string)

	return id == room
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

func getLimitForRoomFrequency(frequency Frequency, base time.Duration) time.Duration {

	// @TODO think about this frequency
	switch frequency {
	case Infrequent:
		return base * 5
	case Normal:
		return base
	case Frequent:
		return base / 2
	}

	return 0
}