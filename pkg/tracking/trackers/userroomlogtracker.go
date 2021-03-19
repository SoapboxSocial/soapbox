package trackers

import (
	"fmt"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/tracking/backends"
)

// UserRoomLogTracker tracks a users join / leave room events.
type UserRoomLogTracker struct {
	backend *backends.UserRoomLogBackend
}

func (r UserRoomLogTracker) CanTrack(event pubsub.Event) bool {
	return event.Type == pubsub.EventTypeRoomLeft
}

func (r UserRoomLogTracker) Track(event pubsub.Event) error {
	if event.Type != pubsub.EventTypeRoomLeft {
		return fmt.Errorf("invalid type for tracker: %d", event.Type)
	}
	panic("implement me")
}
