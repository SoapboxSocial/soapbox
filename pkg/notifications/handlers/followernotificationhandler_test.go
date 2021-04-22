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

func TestFollowerNotificationHandler_Targets(t *testing.T) {
	raw := pubsub.NewFollowerEvent(12, 1)
	event, err := getRawEvent(&raw)
	if err != nil {
		t.Fatal(err)
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	handler := handlers.NewFollowerNotificationHandler(
		notifications.NewTargets(db),
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
		{ID: 1, Frequency: 2, Follows: false},
	}

	if !reflect.DeepEqual(target, expected) {
		t.Fatalf("expected %v actual %v", expected, target)
	}
}

func TestFollowerNotificationHandler_Build(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	handler := handlers.NewFollowerNotificationHandler(notifications.NewTargets(nil), users.NewUserBackend(db))

	displayName := "foo"
	user := 12

	raw := pubsub.NewFollowerEvent(user, 13)

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
		Category: notifications.NEW_FOLLOWER,
		Alert: notifications.Alert{
			Key:       "new_follower_notification",
			Arguments: []string{displayName},
		},
		Arguments: map[string]interface{}{"id": user},
	}

	if !reflect.DeepEqual(n, notification) {
		t.Fatalf("expected %v actual %v", notification, n)
	}
}
