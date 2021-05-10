package handlers_test

import (
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"

	"github.com/soapboxsocial/soapbox/mocks"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/notifications/handlers"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

func TestRoomCreationNotificationHandler_Targets(t *testing.T) {
	raw := pubsub.NewRoomCreationEvent("id", 12, pubsub.Public)
	event, err := getRawEvent(&raw)
	if err != nil {
		t.Fatal(err)
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRoomServiceClient(ctrl)

	handler := handlers.NewRoomCreationNotificationHandler(
		notifications.NewSettings(db),
		nil,
		m,
	)

	mock.
		ExpectPrepare("SELECT").
		ExpectQuery().
		WillReturnRows(mock.NewRows([]string{"user_id", "room_frequency", "follows", "welcome_rooms"}).FromCSVString("1,2,false,false"))

	target, err := handler.Targets(event)
	if err != nil {
		t.Fatal(err)
	}

	expected := []notifications.Target{
		{ID: 1, RoomFrequency: 2, Follows: false, WelcomeRooms: false},
	}

	if !reflect.DeepEqual(target, expected) {
		t.Fatalf("expected %v actual %v", expected, target)
	}
}

func TestRoomCreationNotificationHandler_Build(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockRoomServiceClient(ctrl)

	handler := handlers.NewRoomCreationNotificationHandler(notifications.NewSettings(nil), users.NewBackend(db), m)

	displayName := "foo"
	user := 12
	room := "123"

	raw := pubsub.NewRoomCreationEvent(room, user, pubsub.Public)

	event, err := getRawEvent(&raw)
	if err != nil {
		t.Fatal(err)
	}

	m.EXPECT().GetRoom(gomock.Any(), gomock.Any(), gomock.Any()).Return(&pb.GetRoomResponse{State: &pb.RoomState{Id: "123"}}, nil)

	mock.
		ExpectPrepare("SELECT").
		ExpectQuery().
		WillReturnRows(mock.NewRows([]string{"id", "display_name", "username", "image", "bio", "email"}).FromCSVString("1,foo,t,t,t,t"))

	n, err := handler.Build(event)
	if err != nil {
		t.Fatal(err)
	}

	notification := &notifications.PushNotification{
		Category: notifications.NEW_ROOM,
		Alert: notifications.Alert{
			Key:       "new_room_notification",
			Arguments: []string{displayName},
		},
		Arguments:  map[string]interface{}{"id": room, "creator": user},
		CollapseID: room,
	}

	if !reflect.DeepEqual(n, notification) {
		t.Fatalf("expected %v actual %v", notification, n)
	}
}
