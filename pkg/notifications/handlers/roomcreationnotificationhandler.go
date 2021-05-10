package handlers

import (
	"context"
	"errors"

	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

const testAccountID = 19

type RoomCreationNotificationHandler struct {
	targets  *notifications.Settings
	users    *users.Backend
	metadata pb.RoomServiceClient
}

func NewRoomCreationNotificationHandler(targets *notifications.Settings, u *users.Backend, metadata pb.RoomServiceClient) *RoomCreationNotificationHandler {
	return &RoomCreationNotificationHandler{
		targets:  targets,
		users:    u,
		metadata: metadata,
	}
}

func (r RoomCreationNotificationHandler) Type() pubsub.EventType {
	return pubsub.EventTypeNewRoom
}

func (r RoomCreationNotificationHandler) Origin(event *pubsub.Event) (int, error) {
	creator, err := event.GetInt("creator")
	if err != nil {
		return 0, err
	}

	return creator, nil
}

func (r RoomCreationNotificationHandler) Targets(event *pubsub.Event) ([]notifications.Target, error) {
	if pubsub.RoomVisibility(event.Params["visibility"].(string)) == pubsub.Private {
		return []notifications.Target{}, nil
	}

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

	room := event.Params["id"].(string)
	response, err := r.metadata.GetRoom(context.Background(), &pb.GetRoomRequest{Id: room})
	if err != nil {
		return nil, err
	}

	if response == nil || response.State == nil {
		return nil, errEmptyResponse
	}

	displayName, err := r.getDisplayName(creator)
	if err != nil {
		return nil, err
	}

	if response.State.Name != "" {
		return notifications.NewRoomNotificationWithName(room, displayName, response.State.Name), nil
	}

	return notifications.NewRoomNotification(room, displayName, creator), nil
}

func (r RoomCreationNotificationHandler) getDisplayName(id int) (string, error) {
	user, err := r.users.FindByID(id)
	if err != nil {
		return "", err
	}

	return user.DisplayName, nil
}
