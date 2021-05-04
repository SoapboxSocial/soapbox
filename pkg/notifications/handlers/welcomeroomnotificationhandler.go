package handlers

import (
	"log"

	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

var staticTargets = []notifications.Target{
	{ID: 1},
	{ID: 75},
	{ID: 962},
}

type WelcomeRoomNotificationHandler struct {
	users    *users.UserBackend
	settings *notifications.Settings
}

func NewWelcomeRoomNotificationHandler(u *users.UserBackend, settings *notifications.Settings) *WelcomeRoomNotificationHandler {
	return &WelcomeRoomNotificationHandler{
		users:    u,
		settings: settings,
	}
}

func (w WelcomeRoomNotificationHandler) Type() pubsub.EventType {
	return pubsub.EventTypeWelcomeRoom
}

func (w WelcomeRoomNotificationHandler) Origin(*pubsub.Event) (int, error) {
	return 0, ErrNoCreator
}

func (w WelcomeRoomNotificationHandler) Targets(*pubsub.Event) ([]notifications.Target, error) {
	targets, err := w.settings.GetSettingsForRecentlyActiveUsers()
	if err != nil {
		log.Printf("settings.GetSettingsForRecentlyActiveUsers err: %s", err)
	}

	if len(targets) > 6 {
		targets = targets[:5]
	}

	return append(targets, staticTargets...), nil
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
