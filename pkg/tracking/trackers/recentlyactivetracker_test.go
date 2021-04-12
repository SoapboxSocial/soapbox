package trackers_test

import (
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"

	redisutil "github.com/soapboxsocial/soapbox/pkg/redis"

	"github.com/soapboxsocial/soapbox/pkg/activeusers"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/tracking/trackers"
)

func TestRecentlyActiveTracker_CanTrack(t *testing.T) {
	tests := []pubsub.EventType{
		pubsub.EventTypeUserHeartbeat,
	}

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	tracker := trackers.NewRecentlyActiveTracker(activeusers.NewBackend(db), redisutil.NewTimeoutStore(rdb))

	for _, tt := range tests {
		t.Run(strconv.Itoa(int(tt)), func(t *testing.T) {

			if !tracker.CanTrack(&pubsub.Event{Type: tt}) {
				t.Fatalf("cannot track: %d", tt)
			}
		})
	}
}

func TestRecentlyActiveTracker_Track(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	tracker := trackers.NewRecentlyActiveTracker(activeusers.NewBackend(db), redisutil.NewTimeoutStore(rdb))

	id := 10
	event, err := getRawEvent(pubsub.NewUserHeartbeatEvent(id))
	if err != nil {
		t.Fatal(err)
	}

	mock.
		ExpectPrepare("SELECT update_user_active_times").
		ExpectExec().
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = tracker.Track(event)
	if err != nil {
		t.Fatal(err)
	}

}
