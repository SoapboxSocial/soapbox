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

func NewUserEvent(id int) Event {
	return Event{
		Type:   EventTypeNewUser,
		Params: map[string]interface{}{"id": id},
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
