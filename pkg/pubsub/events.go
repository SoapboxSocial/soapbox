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
		Type: EventTypeUserUpdate,
		Params: map[string]interface{}{"id": id},
	}
}

func NewFollowerEvent(follower, id int) Event {
	return Event{
		Type: EventTypeNewFollower,
		Params: map[string]interface{}{"follower": follower, "id": id},
	}
}