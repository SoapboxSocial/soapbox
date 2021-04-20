package trackers_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/tracking/backends"
	"github.com/soapboxsocial/soapbox/pkg/tracking/trackers"
)

func TestUserRoomLogTracker_Track(t *testing.T) {
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

	tracker := trackers.NewUserRoomLogTracker(
		backends.NewUserRoomLogBackend(db),
		pubsub.NewQueue(rdb),
	)

	event, err := getRawEvent(pubsub.NewRoomLeftEvent("123", 10, pubsub.Public, time.Now()))
	if err != nil {
		t.Fatal(err)
	}

	mock.
		ExpectPrepare("INSERT INTO").
		ExpectExec().
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = tracker.Track(event)
	if err != nil {
		t.Fatal(err)
	}
}
