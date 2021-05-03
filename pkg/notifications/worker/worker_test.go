package worker_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"

	"github.com/soapboxsocial/soapbox/mocks"
	"github.com/soapboxsocial/soapbox/pkg/analytics"
	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/notifications/worker"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
)

func TestWorker(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apns := mocks.NewMockAPNS(ctrl)

	pool := make(chan chan worker.Job)
	w := worker.NewWorker(
		pool,
		&worker.Config{
			APNS:      apns,
			Limiter:   notifications.NewLimiter(rdb, rooms.NewCurrentRoomBackend(db)),
			Devices:   devices.NewBackend(db),
			Store:     notifications.NewStorage(rdb),
			Analytics: analytics.NewBackend(db),
		},
	)

	id := 1
	device := "1234"
	notification := notifications.PushNotification{
		Category:  notifications.ROOM_JOINED,
		Arguments: map[string]interface{}{"creator": 1, "id": "123"},
	}

	mock.
		ExpectPrepare("^SELECT (.+)").
		ExpectQuery().
		WithArgs(id).
		WillReturnRows(mock.NewRows([]string{"room"}).FromCSVString("0"))

	mock.
		ExpectPrepare("^SELECT (.+)").
		ExpectQuery().
		WillReturnRows(mock.NewRows([]string{"token"}).FromCSVString(device))

	apns.EXPECT().Send(gomock.Eq(device), gomock.Any()).Return(nil)

	w.Start()

	queue := <-pool

	queue <- worker.Job{
		Targets:      []notifications.Target{{ID: id, RoomFrequency: notifications.Frequent, Follows: true}},
		Notification: &notification,
	}

	<-pool
}

func TestWorker_WithUnregistered(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	apns := mocks.NewMockAPNS(ctrl)

	pool := make(chan chan worker.Job)
	w := worker.NewWorker(
		pool,
		&worker.Config{
			APNS:      apns,
			Limiter:   notifications.NewLimiter(rdb, rooms.NewCurrentRoomBackend(db)),
			Devices:   devices.NewBackend(db),
			Store:     notifications.NewStorage(rdb),
			Analytics: analytics.NewBackend(db),
		},
	)

	id := 1
	device := "1234"
	notification := notifications.PushNotification{
		Category:  notifications.ROOM_JOINED,
		Arguments: map[string]interface{}{"creator": 1, "id": "123"},
	}

	mock.
		ExpectPrepare("^SELECT (.+)").
		ExpectQuery().
		WithArgs(id).
		WillReturnRows(mock.NewRows([]string{"room"}).FromCSVString("0"))

	mock.
		ExpectPrepare("^SELECT (.+)").
		ExpectQuery().
		WillReturnRows(mock.NewRows([]string{"token"}).FromCSVString(device))

	apns.EXPECT().Send(gomock.Eq(device), gomock.Any()).Return(notifications.ErrDeviceUnregistered)

	mock.
		ExpectPrepare("^DELETE (.+)").
		ExpectExec().
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))

	w.Start()

	queue := <-pool

	queue <- worker.Job{
		Targets:      []notifications.Target{{ID: id, RoomFrequency: notifications.Frequent, Follows: true}},
		Notification: &notification,
	}

	<-pool
}
