package handlers_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"

	"github.com/soapboxsocial/soapbox/mocks"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/notifications/handlers"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

func TestRoomJoinNotificationHandler_Targets(t *testing.T) {
	raw := pubsub.NewRoomJoinEvent("id", 12, pubsub.Public)
	event, err := getRawEvent(&raw)
	if err != nil {
		t.Fatal(err)
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	handler := handlers.NewRoomJoinNotificationHandler(
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

func TestRoomJoinNotificationBuilder_Build(t *testing.T) {
	var tests = []struct {
		event        pubsub.Event
		state        *pb.RoomState
		notification *notifications.PushNotification
	}{
		{
			event: pubsub.NewRoomJoinEvent("xyz", 1, pubsub.Public),
			state: &pb.RoomState{Name: "Test", Members: []*pb.RoomState_RoomMember{
				{DisplayName: "foo"},
			}},
			notification: &notifications.PushNotification{
				Category: notifications.ROOM_JOINED,
				Alert: notifications.Alert{
					Key:       "join_room_name_with_1_notification",
					Arguments: []string{"Test", "foo"},
				},
				Arguments:  map[string]interface{}{"id": "xyz"},
				CollapseID: "xyz",
			},
		},
		{
			event: pubsub.NewRoomJoinEvent("xyz", 1, pubsub.Public),
			state: &pb.RoomState{Members: []*pb.RoomState_RoomMember{
				{DisplayName: "foo"},
			}},
			notification: &notifications.PushNotification{
				Category: notifications.ROOM_JOINED,
				Alert: notifications.Alert{
					Key:       "join_room_with_1_notification",
					Arguments: []string{"foo"},
				},
				Arguments:  map[string]interface{}{"id": "xyz"},
				CollapseID: "xyz",
			},
		},
		{
			event: pubsub.NewRoomJoinEvent("xyz", 1, pubsub.Public),
			state: &pb.RoomState{Name: "Test", Members: []*pb.RoomState_RoomMember{
				{DisplayName: "foo"}, {DisplayName: "bar"},
			}},
			notification: &notifications.PushNotification{
				Category: notifications.ROOM_JOINED,
				Alert: notifications.Alert{
					Key:       "join_room_name_with_2_notification",
					Arguments: []string{"Test", "foo", "bar"},
				},
				Arguments:  map[string]interface{}{"id": "xyz"},
				CollapseID: "xyz",
			},
		},
		{
			event: pubsub.NewRoomJoinEvent("xyz", 1, pubsub.Public),
			state: &pb.RoomState{Name: "Test", Members: []*pb.RoomState_RoomMember{
				{DisplayName: "foo"}, {DisplayName: "bar"}, {DisplayName: "baz"},
			}},
			notification: &notifications.PushNotification{
				Category: notifications.ROOM_JOINED,
				Alert: notifications.Alert{
					Key:       "join_room_name_with_3_notification",
					Arguments: []string{"Test", "foo", "bar", "baz"},
				},
				Arguments:  map[string]interface{}{"id": "xyz"},
				CollapseID: "xyz",
			},
		},
		{
			event: pubsub.NewRoomJoinEvent("xyz", 1, pubsub.Public),
			state: &pb.RoomState{Name: "Test", Members: []*pb.RoomState_RoomMember{
				{DisplayName: "foo", Id: 1}, {DisplayName: "bar"}, {DisplayName: "baz"}, {DisplayName: "bat"},
			}},
			notification: &notifications.PushNotification{
				Category: notifications.ROOM_JOINED,
				Alert: notifications.Alert{
					Key:       "join_room_name_with_3_and_more_notification",
					Arguments: []string{"Test", "foo", "bar", "baz", "1"},
				},
				Arguments:  map[string]interface{}{"id": "xyz"},
				CollapseID: "xyz",
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := mocks.NewMockRoomServiceClient(ctrl)

			handler := handlers.NewRoomJoinNotificationHandler(
				notifications.NewTargets(nil),
				m,
			)

			m.EXPECT().GetRoom(gomock.Any(), gomock.Any(), gomock.Any()).Return(&pb.GetRoomResponse{State: tt.state}, nil)

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
