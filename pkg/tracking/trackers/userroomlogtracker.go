package trackers

import (
	"fmt"
	"time"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/tracking/backends"
)

// UserRoomLogTracker tracks a users join / leave room events.
type UserRoomLogTracker struct {
	backend *backends.UserRoomLogBackend
	queue   *pubsub.Queue
}

func NewUserRoomLogTracker(backend *backends.UserRoomLogBackend, queue *pubsub.Queue) *UserRoomLogTracker {
	return &UserRoomLogTracker{
		backend: backend,
		queue:   queue,
	}
}

func (r UserRoomLogTracker) CanTrack(event *pubsub.Event) bool {
	return event.Type == pubsub.EventTypeRoomLeft
}

func (r UserRoomLogTracker) Track(event *pubsub.Event) error {
	if event.Type != pubsub.EventTypeRoomLeft {
		return fmt.Errorf("invalid type for tracker: %d", event.Type)
	}

	user, err := event.GetInt("creator")
	if err != nil {
		return err
	}

	joined, err := getTime(event, "joined")
	if err != nil {
		return err
	}

	err = r.backend.Store(user, event.Params["id"].(string), event.Params["visibility"].(string), joined, time.Now())
	if err != nil {
		return err
	}

	_ = r.queue.Publish(pubsub.UserTopic, pubsub.NewUserUpdateEvent(user))

	return nil
}

func getTime(event *pubsub.Event, field string) (time.Time, error) {
	value, err := event.GetInt(field)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(int64(value), 0), nil
}
