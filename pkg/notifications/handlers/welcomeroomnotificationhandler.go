package handlers

import (
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type WelcomeRoomNotificationHandler struct {
	users *users.UserBackend
}

func NewWelcomeRoomNotificationHandler(u *users.UserBackend) *WelcomeRoomNotificationHandler {
	return &WelcomeRoomNotificationHandler{
		users: u,
	}
}

func (w WelcomeRoomNotificationHandler) Type() pubsub.EventType {
	return pubsub.EventTypeWelcomeRoom
}

func (w WelcomeRoomNotificationHandler) Origin(*pubsub.Event) (int, error) {
	return 0, ErrNoCreator
}

func (w WelcomeRoomNotificationHandler) Targets(event *pubsub.Event) ([]notifications.Target, error) {
	return []notifications.Target{
		{ID: 1},
		{ID: 75},
		{ID: 962},
		{ID: 13},
		{ID: 6},
	}, nil
}

func (w WelcomeRoomNotificationHandler) Build(event *pubsub.Event) (*notifications.PushNotification, error) {
	creator, err := event.GetInt("id")
	if err != nil {
		return nil, err
	}

	room := event.Params["room"].(string)

	displayName, err := w.getDisplayName(creator)
	if err != nil {
		return nil, err
	}

	return &notifications.PushNotification{
		Category: notifications.WELCOME_ROOM,
		Alert: notifications.Alert{
			Key:       "welcome_room_notification",
			Arguments: []string{displayName},
		},
		Arguments:  map[string]interface{}{"id": room, "from": creator},
		CollapseID: room,
	}, nil
}

func (w WelcomeRoomNotificationHandler) getDisplayName(id int) (string, error) {
	user, err := w.users.FindByID(id)
	if err != nil {
		return "", err
	}

	return user.DisplayName, nil
}
