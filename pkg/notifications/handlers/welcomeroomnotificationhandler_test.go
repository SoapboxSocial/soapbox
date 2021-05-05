package handlers_test

import (
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/notifications/handlers"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

func TestWelcomeRoomNotificationHandler_Targets(t *testing.T) {
	raw := pubsub.NewWelcomeRoomEvent(12, "1234")
	event, err := getRawEvent(&raw)
	if err != nil {
		t.Fatal(err)
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	handler := handlers.NewWelcomeRoomNotificationHandler(
		nil,
		notifications.NewSettings(db),
	)

	mock.
		ExpectPrepare("SELECT").
		ExpectQuery().
		WillReturnRows(mock.NewRows([]string{"user_id", "room_frequency", "follows", "welcome_rooms"}).FromCSVString("12,2,false,false"))

	target, err := handler.Targets(event)
	if err != nil {
		t.Fatal(err)
	}

	expected := []notifications.Target{
		{ID: 12, RoomFrequency: 2, Follows: false, WelcomeRooms: false},
		{ID: 1, RoomFrequency: 0, Follows: false, WelcomeRooms: false},
		{ID: 75, RoomFrequency: 0, Follows: false, WelcomeRooms: false},
		{ID: 962, RoomFrequency: 0, Follows: false, WelcomeRooms: false},
	}

	if !reflect.DeepEqual(target, expected) {
		t.Fatalf("expected %v actual %v", expected, target)
	}
}

func TestWelcomeRoomNotificationHandler_Build(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	handler := handlers.NewWelcomeRoomNotificationHandler(users.NewUserBackend(db), nil)

	displayName := "foo"
	user := 12
	room := "123"

	raw := pubsub.NewWelcomeRoomEvent(12, room)

	event, err := getRawEvent(&raw)
	if err != nil {
		t.Fatal(err)
	}

	mock.
		ExpectPrepare("SELECT").
		ExpectQuery().
		WillReturnRows(mock.NewRows([]string{"id", "display_name", "username", "image", "bio", "email"}).FromCSVString("1,foo,t,t,t,t"))

	n, err := handler.Build(event)
	if err != nil {
		t.Fatal(err)
	}

	notification := &notifications.PushNotification{
		Category: notifications.WELCOME_ROOM,
		Alert: notifications.Alert{
			Key:       "welcome_room_notification",
			Arguments: []string{displayName},
		},
		Arguments:  map[string]interface{}{"id": room, "from": user},
		CollapseID: room,
	}

	if !reflect.DeepEqual(n, notification) {
		t.Fatalf("expected %v actual %v", notification, n)
	}
}
