package handlers_test

import (
	"database/sql/driver"
	"reflect"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/notifications/handlers"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows"
)

func TestFollowRecommendationsNotificationHandler_Targets(t *testing.T) {
	raw := pubsub.NewFollowRecommendationsEvent(12)
	event, err := getRawEvent(&raw)
	if err != nil {
		t.Fatal(err)
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	handler := handlers.NewFollowRecommendationsNotificationHandler(
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

func TestFollowRecommendationsNotificationHandler_Build(t *testing.T) {
	var tests = []struct {
		event        pubsub.Event
		users        [][]driver.Value
		notification *notifications.PushNotification
	}{
		{
			event: pubsub.NewRoomJoinEvent("xyz", 1, pubsub.Public),
			users: [][]driver.Value{
				{1, "bob", "bob", ""},
			},
			notification: &notifications.PushNotification{
				Category: notifications.FOLLOW_RECOMMENDATIONS,
				Alert: notifications.Alert{
					Body:      "bob who you may know is on Soapbox, why not follow them?",
					Key:       "1_follow_recommendations_notification",
					Arguments: []string{"bob"},
				},
			},
		},
		{
			event: pubsub.NewRoomJoinEvent("xyz", 1, pubsub.Public),
			users: [][]driver.Value{
				{1, "bob", "bob", ""},
				{2, "pew", "pew", ""},
			},
			notification: &notifications.PushNotification{
				Category: notifications.FOLLOW_RECOMMENDATIONS,
				Alert: notifications.Alert{
					Body:      "bob and pew who you may know are on Soapbox, why not follow them?",
					Key:       "2_follow_recommendations_notification",
					Arguments: []string{"bob", "pew"},
				},
			},
		},
		{
			event: pubsub.NewRoomJoinEvent("xyz", 1, pubsub.Public),
			users: [][]driver.Value{
				{1, "bob", "bob", ""},
				{2, "pew", "pew", ""},
				{3, "den", "den", ""},
			},
			notification: &notifications.PushNotification{
				Category: notifications.FOLLOW_RECOMMENDATIONS,
				Alert: notifications.Alert{
					Body:      "bob, pew and den who you may know are on Soapbox, why not follow them?",
					Key:       "3_follow_recommendations_notification",
					Arguments: []string{"bob", "pew", "den"},
				},
			},
		},
		{
			event: pubsub.NewRoomJoinEvent("xyz", 1, pubsub.Public),
			users: [][]driver.Value{
				{1, "bob", "bob", ""},
				{2, "pew", "pew", ""},
				{3, "den", "den", ""},
				{4, "n", "n", ""},
			},
			notification: &notifications.PushNotification{
				Category: notifications.FOLLOW_RECOMMENDATIONS,
				Alert: notifications.Alert{
					Body:      "bob, pew, den and 1 others who you may know are on Soapbox, why not follow them?",
					Key:       "3_and_more_follow_recommendations_notification",
					Arguments: []string{"bob", "pew", "den", "1"},
				},
			},
		},
	}

	id := 12
	raw := pubsub.NewFollowRecommendationsEvent(id)
	event, err := getRawEvent(&raw)
	if err != nil {
		t.Fatal(err)
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			handler := handlers.NewFollowRecommendationsNotificationHandler(
				nil,
				follows.NewBackend(db),
			)

			rows := mock.NewRows([]string{"id", "display_name", "username", "image"})
			for _, user := range tt.users {
				rows.AddRow(user...)
			}

			mock.
				ExpectPrepare("^SELECT (.+)").
				ExpectQuery().
				WithArgs(id).
				WillReturnRows(rows)

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
