package pubsub

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

func NewRoomCreationEvent(name string, id, creator int, visibility RoomVisibility) Event {
	return Event{
		Type:   EventTypeNewRoom,
		Params: map[string]interface{}{"name": name, "id": id, "creator": creator, "visibility": visibility},
	}
}


func NewRoomCreationEventWithGroup(name string, id, creator, group int, visibility RoomVisibility) Event {
	return Event{
		Type:   EventTypeNewGroupRoom,
		Params: map[string]interface{}{"name": name, "id": id, "creator": creator, "visibility": visibility, "group": group},
	}
}

func NewRoomJoinEvent(name string, id, creator int, visibility RoomVisibility) Event {
	return Event{
		Type:   EventTypeRoomJoin,
		Params: map[string]interface{}{"name": name, "id": id, "creator": creator, "visibility": visibility},
	}
}

func NewRoomInviteEvent(name string, room, creator, target int) Event {
	return Event{
		Type:   EventTypeRoomInvite,
		Params: map[string]interface{}{"name": name, "room": room, "from": creator, "id": target},
	}
}

func NewRoomLeftEvent(room, user int) Event {
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

func NewGroupInviteEvent(from, id, group int) Event {
	return Event{
		Type:   EventTypeGroupInvite,
		Params: map[string]interface{}{"from": from, "id": id, "group": group},
	}
}
