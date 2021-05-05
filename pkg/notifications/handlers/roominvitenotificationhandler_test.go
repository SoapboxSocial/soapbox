package handlers_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"

	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/notifications/handlers"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

func TestRoomInviteNotificationHandler_Targets(t *testing.T) {
	raw := pubsub.NewRoomInviteEvent("id", "room", 13, 12)
	event, err := getRawEvent(&raw)
	if err != nil {
		t.Fatal(err)
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	handler := handlers.NewRoomInviteNotificationHandler(
		notifications.NewSettings(db),
		nil,
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

func TestRoomInviteNotificationHandler_Build(t *testing.T) {
	var tests = []struct {
		event        pubsub.Event
		notification *notifications.PushNotification
	}{
		{
			event: pubsub.NewRoomInviteEvent("", "xyz", 1, 2),
			notification: &notifications.PushNotification{
				Category: notifications.ROOM_INVITE,
				Alert: notifications.Alert{
					Key:       "room_invite_notification",
					Arguments: []string{"user"},
				},
				Arguments:  map[string]interface{}{"id": "xyz"},
				CollapseID: "xyz",
			},
		},
		{
			event: pubsub.NewRoomInviteEvent("foo", "xyz", 1, 2),
			notification: &notifications.PushNotification{
				Category: notifications.ROOM_INVITE,
				Alert: notifications.Alert{
					Key:       "room_invite_with_name_notification",
					Arguments: []string{"user", "foo"},
				},
				Arguments:  map[string]interface{}{"id": "xyz"},
				CollapseID: "xyz",
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := handlers.NewRoomInviteNotificationHandler(
				notifications.NewSettings(db),
				users.NewUserBackend(db),
			)

			mock.
				ExpectPrepare("SELECT").
				ExpectQuery().
				WillReturnRows(mock.NewRows([]string{"id", "display_name", "username", "image", "bio", "email"}).FromCSVString("1,user,t,t,t,t"))

			event, err := getRawEvent(&tt.event)
			if err != nil {
				t.Fatal(err)
			}

			n, err := handler.Build(event)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(n, tt.notification) {
				t.Fatalf("expected %v actual %v", tt.notification, n)
			}
		})
	}
}
