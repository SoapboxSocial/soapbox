package builders

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang/mock/gomock"

	"github.com/soapboxsocial/soapbox/pkg/followers"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/notifications/builders/internal/mocks"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

func TestRoomJoinNotificationBuilder_Build(t *testing.T) {
	var tests = []struct {
		event        pubsub.Event
		state        *pb.RoomState
		notification *notifications.PushNotification
	}{
		{
			event: pubsub.NewRoomJoinEvent("", "xyz", 1, pubsub.Public),
			state: &pb.RoomState{Name: "Test", Members: []*pb.RoomState_RoomMember{
				{DisplayName: "foo"},
			}},
			notification: &notifications.PushNotification{
				Category: notifications.ROOM_JOINED,
				Alert: notifications.Alert{
					Body:      "Join Test with foo",
					Key:       "join_room_name_with_1",
					Arguments: []string{"Test", "foo"},
				},
				Arguments: map[string]interface{}{"id": "xyz"},
			},
		},
		{
			event: pubsub.NewRoomJoinEvent("", "xyz", 1, pubsub.Public),
			state: &pb.RoomState{Members: []*pb.RoomState_RoomMember{
				{DisplayName: "foo"},
			}},
			notification: &notifications.PushNotification{
				Category: notifications.ROOM_JOINED,
				Alert: notifications.Alert{
					Body:      "Join a room with foo",
					Key:       "join_room_with_1",
					Arguments: []string{"foo"},
				},
				Arguments: map[string]interface{}{"id": "xyz"},
			},
		},
		{
			event: pubsub.NewRoomJoinEvent("", "xyz", 1, pubsub.Public),
			state: &pb.RoomState{Name: "Test", Members: []*pb.RoomState_RoomMember{
				{DisplayName: "foo"}, {DisplayName: "bar"},
			}},
			notification: &notifications.PushNotification{
				Category: notifications.ROOM_JOINED,
				Alert: notifications.Alert{
					Body:      "Join Test with foo and bar",
					Key:       "join_room_name_with_2",
					Arguments: []string{"Test", "foo", "bar"},
				},
				Arguments: map[string]interface{}{"id": "xyz"},
			},
		},
		{
			event: pubsub.NewRoomJoinEvent("", "xyz", 1, pubsub.Public),
			state: &pb.RoomState{Name: "Test", Members: []*pb.RoomState_RoomMember{
				{DisplayName: "foo"}, {DisplayName: "bar"}, {DisplayName: "baz"},
			}},
			notification: &notifications.PushNotification{
				Category: notifications.ROOM_JOINED,
				Alert: notifications.Alert{
					Body:      "Join Test with foo, bar and baz",
					Key:       "join_room_name_with_3",
					Arguments: []string{"Test", "foo", "bar", "baz"},
				},
				Arguments: map[string]interface{}{"id": "xyz"},
			},
		},
		{
			event: pubsub.NewRoomJoinEvent("", "xyz", 1, pubsub.Public),
			state: &pb.RoomState{Name: "Test", Members: []*pb.RoomState_RoomMember{
				{DisplayName: "foo", Id: 1}, {DisplayName: "bar"}, {DisplayName: "baz"}, {DisplayName: "bat"},
			}},
			notification: &notifications.PushNotification{
				Category: notifications.ROOM_JOINED,
				Alert: notifications.Alert{
					Body:      "Join Test with foo, bar, baz and 1 others",
					Key:       "join_room_name_with_3_and_more",
					Arguments: []string{"Test", "foo", "bar", "baz", "1"},
				},
				Arguments: map[string]interface{}{"id": "xyz"},
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

			m := mocks.NewMockRoomServiceClient(ctrl)

			builder := NewRoomJoinNotificationBuilder(
				followers.NewFollowersBackend(db),
				m,
			)

			mock.ExpectPrepare("^SELECT (.+)").ExpectQuery().
				WillReturnRows(mock.NewRows([]string{"follower"}).FromCSVString("2"))

			m.EXPECT().GetRoom(gomock.Any(), gomock.Any(), gomock.Any()).Return(tt.state, nil)

			event, err := getRawEvent(&tt.event)
			if err != nil {
				t.Fatal(err)
			}

			_, n, err := builder.Build(event)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(n, tt.notification) {
				t.Fatalf("expected %v actual %v", tt.notification, n)
			}
		})
	}
}

func getRawEvent(event *pubsub.Event) (*pubsub.Event, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	evt := &pubsub.Event{}
	err = json.Unmarshal(data, evt)
	if err != nil {
		return nil, err
	}

	return evt, nil
}
