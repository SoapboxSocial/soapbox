package notifications

import (
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/rooms"
)

// @TODO MAYBE PUT LIMITS INTO DB? MAKES THEM EASIER TO QUERY.

var (
	followerCooldown     = 15 * time.Minute
	roomInviteCooldown   = 3 * time.Minute
	roomMemberCooldown   = 5 * time.Minute
	roomCooldown         = 10 * time.Minute
	welcomeRoomCooldown  = 30 * time.Minute
	reEngagementCooldown = (24 * time.Hour) * 7
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

func (l *Limiter) ShouldSendNotification(target Target, notification *PushNotification) bool {
	switch notification.Category {
	case NEW_ROOM, ROOM_JOINED:
		if target.RoomFrequency == FrequencyOff {
			return false
		}

		if l.isLimited(limiterKeyForRoom(target.ID, notification)) {
			return false
		}

		if l.isLimited(limiterKeyForRoomMember(target.ID, notification)) {
			return false
		}

		if l.isUserInRoom(target.ID, notification) {
			return false
		}

		return true
	case ROOM_INVITE:
		return !l.isLimited(limiterKeyForRoomInvite(target.ID, notification))
	case NEW_FOLLOWER:
		if !target.Follows {
			return false
		}

		return !l.isLimited(limiterKeyForFollower(target.ID, notification))
	case REENGAGEMENT:
		return !l.isLimited(limiterKeyForFollower(target.ID, notification))
	case WELCOME_ROOM:
		if target.ID == 1 || target.ID == 75 || target.ID == 962 {
			return true
		}

		if !target.WelcomeRooms {
			return false
		}

		return !l.isLimited(limiterKeyForWelcomeRoom(target.ID))
	case TEST:
		return target.ID == 1 || target.ID == 75
	case INFO:
		return true
	default:
		return false
	}
}

func (l *Limiter) SentNotification(target Target, notification *PushNotification) {
	switch notification.Category {
	case NEW_ROOM:
		l.limit(limiterKeyForRoomMember(target.ID, notification), getLimitForRoomFrequency(target.RoomFrequency, roomMemberCooldown))
	case ROOM_JOINED:
		l.limit(limiterKeyForRoomMember(target.ID, notification), getLimitForRoomFrequency(target.RoomFrequency, roomMemberCooldown))
		l.limit(limiterKeyForRoom(target.ID, notification), getLimitForRoomFrequency(target.RoomFrequency, roomCooldown))
	case ROOM_INVITE:
		l.limit(limiterKeyForRoomInvite(target.ID, notification), roomInviteCooldown)
	case NEW_FOLLOWER:
		l.limit(limiterKeyForFollower(target.ID, notification), followerCooldown)
	case REENGAGEMENT:
		l.limit(limiterKeyForReEngagement(target.ID), reEngagementCooldown)
	case WELCOME_ROOM:
		if target.ID == 1 || target.ID == 75 || target.ID == 962 {
			return
		}

		l.limit(limiterKeyForWelcomeRoom(target.ID), welcomeRoomCooldown)
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

func (l *Limiter) isUserInRoom(user int, notification *PushNotification) bool {
	room, _ := l.currentRoom.GetCurrentRoomForUser(user)
	if room == "" {
		return false
	}

	id := notification.Arguments["id"].(string)

	return id == room
}

func limiterKeyForRoom(target int, notification *PushNotification) string {
	return fmt.Sprintf("notifications_limit_%d_room_%v", target, notification.Arguments["id"])
}

func limiterKeyForRoomMember(target int, notification *PushNotification) string {
	return fmt.Sprintf("notifications_limit_%d_room_member_%v", target, notification.Arguments["creator"])
}

func limiterKeyForRoomInvite(target int, notification *PushNotification) string {
	return fmt.Sprintf("notifications_limit_%d_room_invite_%v", target, notification.Arguments["id"])
}

func limiterKeyForFollower(target int, notification *PushNotification) string {
	return fmt.Sprintf("notifications_limit_%d_follower_%v", target, notification.Arguments["id"])
}

func limiterKeyForReEngagement(target int) string {
	return fmt.Sprintf("notifications_limit_%d_re_engagement", target)
}

func limiterKeyForWelcomeRoom(target int) string {
	return fmt.Sprintf("notifications_limit_%d_welcome_room", target)
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
