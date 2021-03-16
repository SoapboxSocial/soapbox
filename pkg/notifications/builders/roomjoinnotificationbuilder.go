package builders

import (
	"context"
	"errors"
	"strconv"

	"github.com/soapboxsocial/soapbox/pkg/followers"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

var (
	errRoomPrivate   = errors.New("room is private")
	errNoRoomMembers = errors.New("room is empty")
	errFailedToSort  = errors.New("failed to sort")
	errEmptyResponse = errors.New("empty response")
)

type RoomJoinNotificationBuilder struct {
	followersBackend *followers.FollowersBackend
	metadata         pb.RoomServiceClient
}

func NewRoomJoinNotificationBuilder(backend *followers.FollowersBackend, metadata pb.RoomServiceClient) *RoomJoinNotificationBuilder {
	return &RoomJoinNotificationBuilder{
		followersBackend: backend,
		metadata:         metadata,
	}
}

func (b *RoomJoinNotificationBuilder) Build(event *pubsub.Event) ([]int, *notifications.PushNotification, error) {
	if pubsub.RoomVisibility(event.Params["visibility"].(string)) == pubsub.Private {
		return nil, nil, errRoomPrivate
	}

	creator, err := event.GetInt("creator")
	if err != nil {
		return nil, nil, err
	}

	targets, err := b.followersBackend.GetAllFollowerIDsFor(creator)
	if err != nil {
		return nil, nil, err
	}

	room := event.Params["id"].(string)
	response, err := b.metadata.GetRoom(context.Background(), &pb.GetRoomRequest{Id: room})
	if err != nil {
		return nil, nil, err
	}

	if response == nil || response.State == nil {
		return nil, nil, errEmptyResponse
	}

	state := response.State

	if len(state.Members) == 0 {
		return nil, nil, errNoRoomMembers
	}

	translation := "join_room_with_"
	body := "Join a room with "
	args := make([]string, 0)

	if state.Name != "" {
		translation = "join_room_name_with_"
		body = "Join \"" + state.Name + "\" with "
		args = append(args, state.Name)
	}

	count := len(state.Members)

	if count == 1 {
		translation += "1"
		body += state.Members[0].DisplayName
		args = append(args, state.Members[0].DisplayName)
	}

	if count == 2 {
		translation += "2"
		body += state.Members[0].DisplayName + " and " + state.Members[1].DisplayName
		args = append(args, state.Members[0].DisplayName, state.Members[1].DisplayName)
	}

	if count == 3 {
		translation += "3"
		body += state.Members[0].DisplayName + ", " + state.Members[1].DisplayName + " and " + state.Members[2].DisplayName
		args = append(args, state.Members[0].DisplayName, state.Members[1].DisplayName, state.Members[2].DisplayName)
	}

	if count > 3 {
		translation += "3_and_more"

		members := members(state.Members, creator)
		if len(members) < 3 {
			return nil, nil, errFailedToSort
		}

		body += members[0] + ", " + members[1] + ", " + members[2] + " and " + strconv.Itoa(count-3) + " others"
		args = append(args, members[0], members[1], members[2], strconv.Itoa(count-3))
	}

	notification := &notifications.PushNotification{
		Category: notifications.ROOM_JOINED,
		Alert: notifications.Alert{
			Body:      body,
			Key:       translation + "_notification",
			Arguments: args,
		},
		CollapseID: room,
		Arguments: map[string]interface{}{"id": room},
	}

	return targets, notification, nil
}

func members(members []*pb.RoomState_RoomMember, first int) []string {
	names := make([]string, 0)

	for i, member := range members {
		if member.Id == int64(first) {
			names = append(names, member.DisplayName)
			members = append(members[:i], members[i+1:]...)
		}
	}

	names = append(names, members[0].DisplayName, members[1].DisplayName)

	return names
}
