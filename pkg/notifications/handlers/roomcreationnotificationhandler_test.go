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

	handler := handlers.NewRoomCreationNotificationHandler(
		notifications.NewSettings(db),
		nil,
	)

	mock.
		ExpectPrepare("SELECT").
		ExpectQuery().
		WillReturnRows(mock.NewRows([]string{"user_id", "room_frequency", "follows"}).FromCSVString("1,2,false"))

	target, err := handler.Targets(event)
	if err != nil {
		t.Fatal(err)
	}

	expected := []notifications.Target{
		{ID: 1, RoomFrequency: 2, Follows: false},
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

	handler := handlers.NewRoomCreationNotificationHandler(notifications.NewSettings(nil), users.NewUserBackend(db))

	displayName := "foo"
	user := 12
	room := "123"

	raw := pubsub.NewRoomCreationEvent(room, user, pubsub.Public)

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
		Category: notifications.NEW_ROOM,
		Alert: notifications.Alert{
			Key:       "new_room_notification",
			Arguments: []string{displayName},
		},
		Arguments:  map[string]interface{}{"id": room},
		CollapseID: room,
	}

	if !reflect.DeepEqual(n, notification) {
		t.Fatalf("expected %v actual %v", notification, n)
	}
}
