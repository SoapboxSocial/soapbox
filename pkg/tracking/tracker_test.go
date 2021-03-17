package tracking_test

import (
	"reflect"
	"testing"

	"github.com/dukex/mixpanel"

	"github.com/soapboxsocial/soapbox/pkg/tracking"
)

func TestMixpanelTracker_Track(t *testing.T) {
	client := mixpanel.NewMock()
	tracker := tracking.NewMixpanelTracker(client)

	id := "123"
	properties := map[string]interface{}{"foo": "bar"}
	event := &tracking.Event{ID: id, Name: tracking.NewUser, Properties: properties}

	err := tracker.Track(event)
	if err != nil {
		t.Fatal(err)
	}

	people := client.People[id]

	if len(people.Events) != 1 {
		t.Fatal("event not logged")
	}

	if !reflect.DeepEqual(people.Events[0].Properties, properties) {
		t.Fatal("did not store properties.")
	}
}
