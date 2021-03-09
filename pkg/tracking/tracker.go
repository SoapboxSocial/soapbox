package tracking

import (
	"github.com/dukex/mixpanel"
)

// Tracker is a interface for tracking Events
type Tracker interface {

	// Track tracks an event, returns an error if failed.
	Track(event Event) error
}

const (
	NewUser = "new_user"
	DeleteUser = "delete_user"
)

type MixpanelTracker struct {
	client mixpanel.Mixpanel
}

func NewMixpanelTracker(client mixpanel.Mixpanel) *MixpanelTracker {
	return &MixpanelTracker{client: client}
}

func (m *MixpanelTracker) Track(event *Event) error {
	if event.Name == DeleteUser {
		return m.client.Update(event.ID, &mixpanel.Update{
			IP:         "0",
			Operation:  "$delete",
		})
	}

	err := m.client.Track(event.ID, event.Name, &mixpanel.Event{IP: "0", Properties: event.Properties})
	if err != nil {
		return err
	}

	if event.Name == NewUser {
		err := m.client.Update(event.ID, &mixpanel.Update{
			IP:         "0",
			Operation:  "$set",
			Properties: event.Properties,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
