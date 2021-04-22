package handlers

import (
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type FollowerNotificationHandler struct {
	targets *notifications.Targets
	users   *users.UserBackend
}

func NewFollowerNotificationHandler(targets *notifications.Targets, u *users.UserBackend) *FollowerNotificationHandler {
	return &FollowerNotificationHandler{
		targets: targets,
		users:   u,
	}
}

func (f FollowerNotificationHandler) Type() pubsub.EventType {
	return pubsub.EventTypeNewFollower
}

func (f FollowerNotificationHandler) Targets(event *pubsub.Event) ([]notifications.Target, error) {
	targetID, err := event.GetInt("id")
	if err != nil {
		return nil, err
	}

	target, err := f.targets.GetTargetFor(targetID)
	if err != nil {
		return nil, err
	}

	return []notifications.Target{*target}, nil
}

func (f FollowerNotificationHandler) Build(event *pubsub.Event) (*notifications.PushNotification, error) {
	creator, err := event.GetInt("follower")
	if err != nil {
		return nil, err
	}

	displayName, err := f.getDisplayName(creator)
	if err != nil {
		return nil, err
	}

	return &notifications.PushNotification{
		Category: notifications.NEW_FOLLOWER,
		Alert: notifications.Alert{
			Key:       "new_follower_notification",
			Arguments: []string{displayName},
		},
		Arguments: map[string]interface{}{"id": creator},
	}, nil
}

func (f FollowerNotificationHandler) getDisplayName(id int) (string, error) {
	user, err := f.users.FindByID(id)
	if err != nil {
		return "", err
	}

	return user.DisplayName, nil
}
