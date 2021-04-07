package pubsub

import (
	"fmt"
	"time"
)

type EventType int

const (
	EventTypeNewRoom EventType = iota
	EventTypeRoomJoin
	EventTypeRoomInvite
	EventTypeNewFollower
	EventTypeUserUpdate
	EventTypeRoomLeft
	EventTypeNewUser
	EventTypeNewStory
	EventTypeStoryReaction
	EventTypeUserHeartbeat
	EventTypeWelcomeRoom
	EventTypeRoomLinkShare
	EventTypeRoomOpenMini
	EventTypeDeleteUser
)

type RoomVisibility string

const (
	Public  RoomVisibility = "public"
	Private RoomVisibility = "private"
)

type Event struct {
	Type   EventType              `json:"type"`
	Params map[string]interface{} `json:"params"`
}

func (e Event) GetInt(field string) (int, error) {
	val, ok := e.Params[field].(float64)
	if !ok {
		return 0, fmt.Errorf("failed to recover %s", field)
	}

	return int(val), nil
}

func NewUserUpdateEvent(id int) Event {
	return Event{
		Type:   EventTypeUserUpdate,
		Params: map[string]interface{}{"id": id},
	}
}

func NewUserEvent(id int, username string) Event {
	return Event{
		Type:   EventTypeNewUser,
		Params: map[string]interface{}{"id": id, "username": username},
	}
}

func NewFollowerEvent(follower, id int) Event {
	return Event{
		Type:   EventTypeNewFollower,
		Params: map[string]interface{}{"follower": follower, "id": id},
	}
}

func NewRoomCreationEvent(id string, creator int, visibility RoomVisibility) Event {
	return Event{
		Type:   EventTypeNewRoom,
		Params: map[string]interface{}{"id": id, "creator": creator, "visibility": visibility},
	}
}

func NewRoomJoinEvent(id string, creator int, visibility RoomVisibility) Event {
	return Event{
		Type:   EventTypeRoomJoin,
		Params: map[string]interface{}{"id": id, "creator": creator, "visibility": visibility},
	}
}

func NewStoryCreationEvent(creator int) Event {
	return Event{
		Type:   EventTypeNewStory,
		Params: map[string]interface{}{"creator": creator},
	}
}

func NewStoryReactionEvent(user int) Event {
	return Event{
		Type:   EventTypeStoryReaction,
		Params: map[string]interface{}{"id": user},
	}
}

func NewRoomInviteEvent(name, room string, creator, target int) Event {
	return Event{
		Type:   EventTypeRoomInvite,
		Params: map[string]interface{}{"name": name, "room": room, "from": creator, "id": target},
	}
}

func NewRoomLeftEvent(room string, user int, visibility RoomVisibility, joined time.Time) Event {
	return Event{
		Type:   EventTypeRoomLeft,
		Params: map[string]interface{}{"id": room, "creator": user, "joined": joined.Unix(), "visibility": visibility},
	}
}

func NewUserHeartbeatEvent(user int) Event {
	return Event{
		Type:   EventTypeUserHeartbeat,
		Params: map[string]interface{}{"id": user},
	}
}

func NewWelcomeRoomEvent(user int, room string) Event {
	return Event{
		Type:   EventTypeWelcomeRoom,
		Params: map[string]interface{}{"id": user, "room": room},
	}
}

func NewRoomLinkShareEvent(user int, room string) Event {
	return Event{
		Type:   EventTypeRoomLinkShare,
		Params: map[string]interface{}{"id": user, "room": room},
	}
}

func NewRoomOpenMiniEvent(user, mini int, room string) Event {
	return Event{
		Type:   EventTypeRoomOpenMini,
		Params: map[string]interface{}{"id": user, "mini": mini, "room": room},
	}
}

func NewDeleteUserEvent(user int) Event {
	return Event{
		Type:   EventTypeDeleteUser,
		Params: map[string]interface{}{"id": user},
	}
}
