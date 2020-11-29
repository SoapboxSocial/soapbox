package notifications

import "fmt"

type NotificationCategory string

const (
	NEW_ROOM     NotificationCategory = "NEW_ROOM"
	NEW_FOLLOWER NotificationCategory = "NEW_FOLLOWER"
	ROOM_INVITE  NotificationCategory = "ROOM_INVITE"
	ROOM_JOINED  NotificationCategory = "ROOM_JOINED"
	GROUP_INVITE NotificationCategory = "GROUP_INVITE"
)

type Alert struct {
	Body      string   `json:"body,omitempty"`
	Key       string   `json:"loc-key"`
	Arguments []string `json:"loc-args"`
}

// PushNotification is JSON encoded and sent to the APNS service.
type PushNotification struct {
	Category  NotificationCategory   `json:"category"`
	Alert     Alert                  `json:"alert"`
	Arguments map[string]interface{} `json:"arguments"`
}

// Notification is stored in redis for the notification endpoint.
type Notification struct {
	Timestamp int64                  `json:"timestamp"`
	From      int                    `json:"from"`
	Category  NotificationCategory   `json:"category"`
	Arguments map[string]interface{} `json:"arguments"`
}

func NewRoomNotification(id int, creator string) *PushNotification {
	return &PushNotification{
		Category: NEW_ROOM,
		Alert: Alert{
			Key:       "new_room_notification",
			Arguments: []string{creator},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomNotificationWithName(id int, creator, name string) *PushNotification {
	return &PushNotification{
		Category: NEW_ROOM,
		Alert: Alert{
			Key:       "new_room_with_name_notification",
			Arguments: []string{creator, name},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomWithGroupNotification(id int, creator, group string) *PushNotification {
	return &PushNotification{
		Category: NEW_ROOM,
		Alert: Alert{
			Key:       "new_room_with_group_notification",
			Body:      fmt.Sprintf("%s created a room in \"%s\", why not join them?", creator, group),
			Arguments: []string{creator, group},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomWithGroupAndNameNotification(id int, creator, group, name string) *PushNotification {
	return &PushNotification{
		Category: NEW_ROOM,
		Alert: Alert{
			Key:       "new_room_with_group_and_name_notification",
			Body:      fmt.Sprintf("%s created the room \"%s\" in \"%s\", why not join them?", creator, name, group),
			Arguments: []string{creator, name, group},
		},
		Arguments: map[string]interface{}{"id": id},
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

func NewRoomJoinedNotification(id int, participant string) *PushNotification {
	return &PushNotification{
		Category: ROOM_JOINED,
		Alert: Alert{
			Key:       "room_joined_notification",
			Arguments: []string{participant},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomJoinedNotificationWithName(id int, participant, name string) *PushNotification {
	return &PushNotification{
		Category: ROOM_JOINED,
		Alert: Alert{
			Key:       "room_joined_with_name_notification",
			Arguments: []string{participant, name},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomInviteNotification(id int, from string) *PushNotification {
	return &PushNotification{
		Category: ROOM_INVITE,
		Alert: Alert{
			Key:       "room_invite_notification",
			Arguments: []string{from},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomInviteNotificationWithName(id int, from, room string) *PushNotification {
	return &PushNotification{
		Category: ROOM_JOINED,
		Alert: Alert{
			Key:       "room_invite_with_name_notification",
			Arguments: []string{from, room},
		},
		Arguments: map[string]interface{}{"id": id},
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
