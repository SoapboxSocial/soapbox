package notifications_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"

	"github.com/soapboxsocial/soapbox/mocks"
	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
)

func TestService_Send(t *testing.T) {
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

	service := notifications.NewService(
		apns,
		notifications.NewLimiter(rdb, rooms.NewCurrentRoomBackend(db)),
		devices.NewBackend(db),
		notifications.NewStorage(rdb),
	)

	id := 1
	device := "1234"
	notification := notifications.PushNotification{Category: notifications.ROOM_JOINED}

	mock.
		ExpectPrepare("^SELECT (.+)").
		ExpectQuery().
		WithArgs(id).
		WillReturnRows(mock.NewRows([]string{"room"}).FromCSVString("0"))

	mock.
		ExpectPrepare("^SELECT (.+)").
		ExpectQuery().
		WithArgs(id).
		WillReturnRows(mock.NewRows([]string{"token"}).FromCSVString(device))

	apns.EXPECT().Send(gomock.Eq(device), gomock.Any()).Return(nil)

	event := pubsub.NewRoomJoinEvent("123", 2, pubsub.Public)
	service.Send(
		notifications.Target{ID: id, RoomFrequency: notifications.Frequent, Follows: true},
		&event,
		&notification,
	)
}
