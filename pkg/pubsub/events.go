package pubsub

import (
	"fmt"
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
	EventTypeNewGroup
	EventTypeGroupInvite
	EventTypeNewGroupRoom
	EventTypeGroupUpdate
	EventTypeGroupJoin
	EventTypeGroupDelete
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

func NewRoomCreationEvent(name, id string, creator int, visibility RoomVisibility) Event {
	return Event{
		Type:   EventTypeNewRoom,
		Params: map[string]interface{}{"name": name, "id": id, "creator": creator, "visibility": visibility},
	}
}

func NewRoomCreationEventWithGroup(name, id string, creator, group int, visibility RoomVisibility) Event {
	return Event{
		Type:   EventTypeNewGroupRoom,
		Params: map[string]interface{}{"name": name, "id": id, "creator": creator, "visibility": visibility, "group": group},
	}
}

func NewRoomJoinEvent(name, id string, creator int, visibility RoomVisibility) Event {
	return Event{
		Type:   EventTypeRoomJoin,
		Params: map[string]interface{}{"name": name, "id": id, "creator": creator, "visibility": visibility},
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

func NewRoomLeftEvent(room string, user int) Event {
	return Event{
		Type:   EventTypeRoomLeft,
		Params: map[string]interface{}{"id": room, "creator": user},
	}
}

func NewGroupCreationEvent(id, creator int, name string) Event {
	return Event{
		Type:   EventTypeNewGroup,
		Params: map[string]interface{}{"id": id, "creator": creator, "name": name},
	}
}

func NewGroupUpdateEvent(id int) Event {
	return Event{
		Type:   EventTypeGroupUpdate,
		Params: map[string]interface{}{"id": id},
	}
}

func NewGroupInviteEvent(from, id, group int) Event {
	return Event{
		Type:   EventTypeGroupInvite,
		Params: map[string]interface{}{"from": from, "id": id, "group": group},
	}
}

func NewGroupJoinEvent(id, group int) Event {
	return Event{
		Type:   EventTypeGroupJoin,
		Params: map[string]interface{}{"id": id, "group": group},
	}
}

func NewGroupDeleteEvent(group int) Event {
	return Event{
		Type:   EventTypeGroupDelete,
		Params: map[string]interface{}{"group": group},
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

func NewRoomOpenMiniEvent(user int, mini, room string) Event {
	return Event{
		Type:   EventTypeRoomOpenMini,
		Params: map[string]interface{}{"id": user, "mini": mini, "room": room},
	}
}

func NewDeleteUserEvent(user int,) Event {
	return Event{
		Type:   EventTypeDeleteUser,
		Params: map[string]interface{}{"id": user},
	}
}
