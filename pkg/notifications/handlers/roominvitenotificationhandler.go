package handlers

import (
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type RoomInviteNotificationHandler struct {
	targets *notifications.Settings
	users   *users.UserBackend
}

func NewRoomInviteNotificationHandler(targets *notifications.Settings, u *users.UserBackend) *RoomInviteNotificationHandler {
	return &RoomInviteNotificationHandler{
		targets: targets,
		users:   u,
	}
}

func (r RoomInviteNotificationHandler) Type() pubsub.EventType {
	return pubsub.EventTypeRoomInvite
}

func (r RoomInviteNotificationHandler) Targets(event *pubsub.Event) ([]notifications.Target, error) {
	targetID, err := event.GetInt("id")
	if err != nil {
		return nil, err
	}

	target, err := r.targets.GetSettingsFor(targetID)
	if err != nil {
		return nil, err
	}

	return []notifications.Target{*target}, nil
}

func (r RoomInviteNotificationHandler) Build(event *pubsub.Event) (*notifications.PushNotification, error) {
	creator, err := event.GetInt("from")
	if err != nil {
		return nil, err
	}

	name := event.Params["name"].(string)
	room := event.Params["room"].(string)

	displayName, err := r.getDisplayName(creator)
	if err != nil {
		return nil, err
	}

	if name == "" {
		return notifications.NewRoomInviteNotification(room, displayName), nil
	}

	return notifications.NewRoomInviteNotificationWithName(room, displayName, name), nil
}

func (r RoomInviteNotificationHandler) getDisplayName(id int) (string, error) {
	user, err := r.users.FindByID(id)
	if err != nil {
		return "", err
	}

	return user.DisplayName, nil
}
