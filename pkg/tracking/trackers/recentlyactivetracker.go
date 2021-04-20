package trackers

import (
	"fmt"
	"log"
	"time"

	"github.com/soapboxsocial/soapbox/pkg/activeusers"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/redis"
)

type RecentlyActiveTracker struct {
	backend *activeusers.Backend
	timeout *redis.TimeoutStore
}

func NewRecentlyActiveTracker(backend *activeusers.Backend, timeout *redis.TimeoutStore) *RecentlyActiveTracker {
	return &RecentlyActiveTracker{
		backend: backend,
		timeout: timeout,
	}
}

func (r *RecentlyActiveTracker) CanTrack(event *pubsub.Event) bool {
	return event.Type == pubsub.EventTypeUserHeartbeat ||
		event.Type == pubsub.EventTypeRoomLeft
}

func (r *RecentlyActiveTracker) Track(event *pubsub.Event) error {
	id, err := event.GetInt("id")
	if err != nil {
		return err
	}

	timeoutkey := fmt.Sprintf("recently_active_timout_%d", id)
	if r.timeout.IsOnTimeout(timeoutkey) {
		return nil
	}

	err = r.backend.SetLastActiveTime(id, time.Now())
	if err != nil {
		return err
	}

	err = r.timeout.SetTimeout(timeoutkey, 3*time.Minute)
	if err != nil {
		log.Printf("failed to set recently active timeout err: %s", err)
	}

	return nil
}
