package trackers

import (
	"fmt"
	"strconv"

	"github.com/dukex/mixpanel"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/tracking"
)

const (
	NewUser    = "new_user"
	DeleteUser = "delete_user"
)

type MixpanelTracker struct {
	client mixpanel.Mixpanel
}

func NewMixpanelTracker(client mixpanel.Mixpanel) *MixpanelTracker {
	return &MixpanelTracker{client: client}
}

func (m *MixpanelTracker) CanTrack(event *pubsub.Event) bool {
	return event.Type != pubsub.EventTypeRoomInvite &&
		event.Type != pubsub.EventTypeUserUpdate &&
		event.Type != pubsub.EventTypeGroupInvite &&
		event.Type != pubsub.EventTypeGroupUpdate &&
		event.Type != pubsub.EventTypeGroupDelete &&
		event.Type != pubsub.EventTypeWelcomeRoom
}

func (m *MixpanelTracker) Track(event *pubsub.Event) error {
	log := transform(event)
	if log == nil {
		return fmt.Errorf("invalid type for tracker: %d", event.Type)
	}

	if log.Name == DeleteUser {
		return m.client.Update(log.ID, &mixpanel.Update{
			IP:        "0",
			Operation: "$delete",
		})
	}

	err := m.client.Track(log.ID, log.Name, &mixpanel.Event{IP: "0", Properties: log.Properties})
	if err != nil {
		return err
	}

	if log.Name == NewUser {
		err := m.client.Update(log.ID, &mixpanel.Update{
			IP:         "0",
			Operation:  "$set",
			Properties: log.Properties,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func transform(event *pubsub.Event) *tracking.Event {
	switch event.Type {
	case pubsub.EventTypeNewRoom:

		id, err := event.GetInt("creator")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "room_new",
			Properties: map[string]interface{}{
				"room_id":    event.Params["id"],
				"visibility": event.Params["visibility"],
			},
		}
	case pubsub.EventTypeRoomJoin:
		id, err := event.GetInt("creator")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "room_join",
			Properties: map[string]interface{}{
				"room_id":    event.Params["id"],
				"visibility": event.Params["visibility"],
			},
		}
	case pubsub.EventTypeRoomLeft:
		id, err := event.GetInt("creator")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "room_left",
			Properties: map[string]interface{}{
				"room_id": event.Params["id"],
			},
		}
	case pubsub.EventTypeNewUser:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: NewUser,
			Properties: map[string]interface{}{
				"user_id":  event.Params["id"],
				"username": event.Params["username"],
			},
		}
	case pubsub.EventTypeNewGroup:
		id, err := event.GetInt("creator")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "group_new",
			Properties: map[string]interface{}{
				"group_id": event.Params["id"],
				"name":     event.Params["name"],
			},
		}
	case pubsub.EventTypeNewGroupRoom:
		id, err := event.GetInt("creator")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "room_new",
			Properties: map[string]interface{}{
				"room_id":    event.Params["id"],
				"visibility": event.Params["visibility"],
				"group_id":   event.Params["group"],
			},
		}
	case pubsub.EventTypeNewFollower:
		id, err := event.GetInt("follower")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "followed",
			Properties: map[string]interface{}{
				"following_id": event.Params["id"],
			},
		}
	case pubsub.EventTypeGroupJoin:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "group_join",
			Properties: map[string]interface{}{
				"group": event.Params["group"],
			},
		}
	case pubsub.EventTypeNewStory:
		id, err := event.GetInt("creator")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:         strconv.Itoa(id),
			Name:       "story_new",
			Properties: map[string]interface{}{},
		}
	case pubsub.EventTypeStoryReaction:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:         strconv.Itoa(id),
			Name:       "story_reaction",
			Properties: map[string]interface{}{},
		}
	case pubsub.EventTypeUserHeartbeat:

		// @TODO ADD HEARTBEAT COOLDOWN of 30 mins

		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:         strconv.Itoa(id),
			Name:       "heartbeat",
			Properties: map[string]interface{}{},
		}
	case pubsub.EventTypeRoomOpenMini:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		mini, err := event.GetInt("mini")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "room_open_mini",
			Properties: map[string]interface{}{
				"room_id": event.Params["room"],
				"mini":    mini,
			},
		}
	case pubsub.EventTypeRoomLinkShare:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: "room_link_share",
			Properties: map[string]interface{}{
				"room_id": event.Params["room"],
			},
		}
	case pubsub.EventTypeDeleteUser:
		id, err := event.GetInt("id")
		if err != nil {
			return nil
		}

		return &tracking.Event{
			ID:   strconv.Itoa(id),
			Name: DeleteUser,
		}
	}

	return nil
}
