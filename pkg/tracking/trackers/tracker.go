package trackers

import "github.com/soapboxsocial/soapbox/pkg/pubsub"

// Tracker is a interface for tracking Events
type Tracker interface {

	// CanTrack returns whether tracker tracks a specific event.
	CanTrack(event pubsub.Event) bool

	// Track tracks an event, returns an error if failed.
	Track(event pubsub.Event) error
}

