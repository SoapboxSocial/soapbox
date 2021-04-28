package handlers

import (
	"errors"

	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

const testAccountID = 19

type RoomCreationNotificationHandler struct {
	targets *notifications.Settings
	users   *users.UserBackend
}

func NewRoomCreationNotificationHandler(targets *notifications.Settings, u *users.UserBackend) *RoomCreationNotificationHandler {
	return &RoomCreationNotificationHandler{
		targets: targets,
		users:   u,
	}
}

func (r RoomCreationNotificationHandler) Type() pubsub.EventType {
	return pubsub.EventTypeNewRoom
}

func (r RoomCreationNotificationHandler) Targets(event *pubsub.Event) ([]notifications.Target, error) {
	creator, err := event.GetInt("creator")
	if err != nil {
		return nil, err
	}

	targets, err := r.targets.GetSettingsFollowingUser(creator)
	if err != nil {
		return nil, err
	}

	return targets, nil
}

func (r RoomCreationNotificationHandler) Build(event *pubsub.Event) (*notifications.PushNotification, error) {
	if pubsub.RoomVisibility(event.Params["visibility"].(string)) == pubsub.Private {
		return nil, errors.New("room is private")
	}

	creator, err := event.GetInt("creator")
	if err != nil {
		return nil, err
	}

	if creator == testAccountID {
		return nil, errors.New("test account started room")
	}

	// Quick fix
	//name := event.Params["name"].(string)
	room := event.Params["id"].(string)

	displayName, err := r.getDisplayName(creator)
	if err != nil {
		return nil, err
	}

	notification := func() *notifications.PushNotification {
		//if name == "" {
		return notifications.NewRoomNotification(room, displayName)
		//}

		//return notifications.NewRoomNotificationWithName(room, displayName, name)
	}()

	return notification, nil
}

func (r RoomCreationNotificationHandler) getDisplayName(id int) (string, error) {
	user, err := r.users.FindByID(id)
	if err != nil {
		return "", err
	}

	return user.DisplayName, nil
}
