package trackers

import (
	"github.com/soapboxsocial/soapbox/pkg/activeusers"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
)

type RecentlyActiveTracker struct {
	backend *activeusers.Backend
}

func NewRecentlyActiveTracker(backend *activeusers.Backend) *RecentlyActiveTracker {
	return &RecentlyActiveTracker{
		backend: backend,
	}
}

func (r *RecentlyActiveTracker) CanTrack(event *pubsub.Event) bool {
	return event.Type == pubsub.EventTypeUserHeartbeat
}

func (r *RecentlyActiveTracker) Track(event *pubsub.Event) error {
	panic("implement me")
}
