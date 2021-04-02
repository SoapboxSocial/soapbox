package trackers

import (
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
)

type RecentlyActiveTracker struct {
}

func (r *RecentlyActiveTracker) CanTrack(event *pubsub.Event) bool {
	return event.Type == pubsub.EventTypeUserHeartbeat
}

func (r *RecentlyActiveTracker) Track(event *pubsub.Event) error {
	panic("implement me")
}
