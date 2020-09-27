package pubsub

type EventType int

const (
	EventTypeNewRoom EventType = iota
	EventTypeRoomJoin
	EventTypeRoomInvite
	EventTypeNewFollower
	EventTypeUserUpdate
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

func NewFollowerEvent(follower, id int) Event {
	return Event{
		Type:   EventTypeNewFollower,
		Params: map[string]interface{}{"follower": follower, "id": id},
	}
}

func NewRoomCreationEvent(name string, id, creator int) Event {
	return Event{
		Type:   EventTypeNewRoom,
		Params: map[string]interface{}{"name": name, "id": id, "creator": creator},
	}
}

func NewRoomJoinEvent(name string, id, creator int) Event {
	return Event{
		Type:   EventTypeRoomJoin,
		Params: map[string]interface{}{"name": name, "id": id, "creator": creator},
	}
}

func NewRoomInviteEvent(name string, room, creator, target int) Event {
	return Event{
		Type:   EventTypeRoomInvite,
		Params: map[string]interface{}{"name": name, "room": room, "from": creator, "id": target},
	}
}
