package notifications

import "fmt"

type NotificationCategory string

const (
	NEW_ROOM     NotificationCategory = "NEW_ROOM"
	NEW_FOLLOWER NotificationCategory = "NEW_FOLLOWER"
	ROOM_INVITE  NotificationCategory = "ROOM_INVITE"
	ROOM_JOINED  NotificationCategory = "ROOM_JOINED"
	GROUP_INVITE NotificationCategory = "GROUP_INVITE"
	WELCOME_ROOM NotificationCategory = "WELCOME_ROOM"
)

type Alert struct {
	Body      string   `json:"body,omitempty"`
	Key       string   `json:"loc-key"`
	Arguments []string `json:"loc-args"`
}

// PushNotification is JSON encoded and sent to the APNS service.
type PushNotification struct {
	Category   NotificationCategory   `json:"category"`
	Alert      Alert                  `json:"alert"`
	Arguments  map[string]interface{} `json:"arguments"`
	CollapseID string                 `json:"-"`
}

// Notification is stored in redis for the notification endpoint.
type Notification struct {
	Timestamp int64                  `json:"timestamp"`
	From      int                    `json:"from"`
	Category  NotificationCategory   `json:"category"`
	Arguments map[string]interface{} `json:"arguments"`
}

func NewRoomNotification(id, creator string) *PushNotification {
	return &PushNotification{
		Category: NEW_ROOM,
		Alert: Alert{
			Key:       "new_room_notification",
			Arguments: []string{creator},
		},
		Arguments:  map[string]interface{}{"id": id},
		CollapseID: id,
	}
}

func NewRoomNotificationWithName(id, creator, name string) *PushNotification {
	return &PushNotification{
		Category: NEW_ROOM,
		Alert: Alert{
			Key:       "new_room_with_name_notification",
			Arguments: []string{creator, name},
		},
		Arguments:  map[string]interface{}{"id": id},
		CollapseID: id,
	}
}

func NewRoomWithGroupNotification(id, creator, group string) *PushNotification {
	return &PushNotification{
		Category: NEW_ROOM,
		Alert: Alert{
			Key:       "new_room_with_group_notification",
			Body:      fmt.Sprintf("%s created a room in \"%s\", why not join them?", creator, group),
			Arguments: []string{creator, group},
		},
		Arguments:  map[string]interface{}{"id": id},
		CollapseID: id,
	}
}

func NewRoomWithGroupAndNameNotification(id, creator, group, name string) *PushNotification {
	return &PushNotification{
		Category: NEW_ROOM,
		Alert: Alert{
			Key:       "new_room_with_group_and_name_notification",
			Body:      fmt.Sprintf("%s created the room \"%s\" in \"%s\", why not join them?", creator, name, group),
			Arguments: []string{creator, name, group},
		},
		Arguments:  map[string]interface{}{"id": id},
		CollapseID: id,
	}
}

func NewFollowerNotification(id int, follower string) *PushNotification {
	return &PushNotification{
		Category: NEW_FOLLOWER,
		Alert: Alert{
			Key:       "new_follower_notification",
			Arguments: []string{follower},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomJoinedNotification(id, participant string) *PushNotification {
	return &PushNotification{
		Category: ROOM_JOINED,
		Alert: Alert{
			Key:       "room_joined_notification",
			Arguments: []string{participant},
		},
		Arguments:  map[string]interface{}{"id": id},
		CollapseID: id,
	}
}

func NewRoomJoinedNotificationWithName(id, participant, name string) *PushNotification {
	return &PushNotification{
		Category: ROOM_JOINED,
		Alert: Alert{
			Key:       "room_joined_with_name_notification",
			Arguments: []string{participant, name},
		},
		Arguments:  map[string]interface{}{"id": id},
		CollapseID: id,
	}
}

func NewRoomInviteNotification(id, from string) *PushNotification {
	return &PushNotification{
		Category: ROOM_INVITE,
		Alert: Alert{
			Key:       "room_invite_notification",
			Arguments: []string{from},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomInviteNotificationWithName(id, from, room string) *PushNotification {
	return &PushNotification{
		Category: ROOM_JOINED,
		Alert: Alert{
			Key:       "room_invite_with_name_notification",
			Arguments: []string{from, room},
		},
		Arguments:  map[string]interface{}{"id": id},
		CollapseID: room,
	}
}

func NewGroupInviteNotification(groupId, fromId int, from, group string) *PushNotification {
	return &PushNotification{
		Category: GROUP_INVITE,
		Alert: Alert{
			Key:       "group_invite_notification",
			Arguments: []string{from, group},
		},
		Arguments: map[string]interface{}{"id": groupId, "from": fromId},
	}
}

func NewWelcomeRoomNotification(user, room string, from int) *PushNotification {
	return &PushNotification{
		Category: WELCOME_ROOM,
		Alert: Alert{
			Key:       "welcome_room_notification",
			Body:      fmt.Sprintf("%s just signed up, why not welcome them?", user),
			Arguments: []string{user},
		},
		Arguments:  map[string]interface{}{"id": room, "from": from},
		CollapseID: room,
	}
}
