package trackers_test

import (
	"encoding/json"
	"reflect"
	"strconv"
	"testing"

	"github.com/dukex/mixpanel"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/tracking/trackers"
)

func TestMixpanelTracker_Track(t *testing.T) {
	client := mixpanel.NewMock()
	tracker := trackers.NewMixpanelTracker(client)

	id := 123
	event, err := getRawEvent(pubsub.NewUserEvent(id, "foo"))
	if err != nil {
		t.Fatal(err)
	}

	err = tracker.Track(event)
	if err != nil {
		t.Fatal(err)
	}

	people := client.People[strconv.Itoa(id)]

	if len(people.Events) != 1 {
		t.Fatal("event not logged")
	}

	if !reflect.DeepEqual(people.Events[0].Properties, map[string]interface{}{"user_id": 123.0, "username": "foo"}) {
		t.Fatal("did not store properties.")
	}
}

func TestMixpanelTracker_CanTrack(t *testing.T) {
	tests := []pubsub.EventType{
		pubsub.EventTypeNewRoom,
		pubsub.EventTypeRoomJoin,
		pubsub.EventTypeNewFollower,
		pubsub.EventTypeRoomLeft,
		pubsub.EventTypeNewUser,
		pubsub.EventTypeNewGroup,
		pubsub.EventTypeGroupJoin,
		pubsub.EventTypeNewGroupRoom,
		pubsub.EventTypeNewStory,
		pubsub.EventTypeStoryReaction,
		pubsub.EventTypeUserHeartbeat,
		pubsub.EventTypeRoomLinkShare,
		pubsub.EventTypeRoomOpenMini,
		pubsub.EventTypeDeleteUser,
	}

	client := mixpanel.NewMock()
	tracker := trackers.NewMixpanelTracker(client)

	for _, tt := range tests {
		t.Run(strconv.Itoa(int(tt)), func(t *testing.T) {

			if !tracker.CanTrack(&pubsub.Event{Type: tt}) {
				t.Fatalf("cannot track: %d", tt)
			}
		})
	}
}

func getRawEvent(event pubsub.Event) (*pubsub.Event, error) {
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
