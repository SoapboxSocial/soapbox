package notifications

import "fmt"

type NotificationCategory string

const (
	NEW_ROOM     NotificationCategory = "NEW_ROOM"
	NEW_FOLLOWER NotificationCategory = "NEW_FOLLOWER"
	ROOM_INVITE  NotificationCategory = "ROOM_INVITE"
	ROOM_JOINED  NotificationCategory = "ROOM_JOINED"
)

type Alert struct {
	Body      string   `json:"body,omitempty"`
	Key       string   `json:"loc-key"`
	Arguments []string `json:"loc-args"`
}

type Notification struct {
	Category  NotificationCategory   `json:"category"`
	Alert     Alert                  `json:"alert"`
	Arguments map[string]interface{} `json:"arguments"`
}

func NewRoomNotification(id int, creator string) *Notification {
	return &Notification{
		Category: NEW_ROOM,
		Alert: Alert{
			Key:       "new_room_notification",
			Arguments: []string{creator},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomNotificationWithName(id int, creator, name string) *Notification {
	return &Notification{
		Category: NEW_ROOM,
		Alert: Alert{
			Key:       "new_room_with_name_notification",
			Arguments: []string{creator, name},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewFollowerNotification(id int, follower string) *Notification {
	return &Notification{
		Category: NEW_FOLLOWER,
		Alert: Alert{
			Key:       "new_follower_notification",
			Arguments: []string{follower},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomJoinedNotification(id int, participant string) *Notification {
	return &Notification{
		Category: ROOM_JOINED,
		Alert: Alert{
			Body:      fmt.Sprintf("%s joined a room, why not join them?", participant),
			Key:       "room_joined_notification",
			Arguments: []string{participant},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomJoinedNotificationWithName(id int, participant, name string) *Notification {
	return &Notification{
		Category: ROOM_JOINED,
		Alert: Alert{
			Body:      fmt.Sprintf("%s joined the room \"%s\", why not join them?", participant, name),
			Key:       "room_joined_with_name_notification",
			Arguments: []string{participant, name},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomInviteNotification(id int, from string) *Notification {
	return &Notification{
		Category: ROOM_INVITE,
		Alert: Alert{
			Body:      fmt.Sprintf("%s invited you to join a room. why not join them?", from),
			Key:       "room_invite_notification",
			Arguments: []string{from},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}

func NewRoomInviteNotificationWithName(id int, from, room string) *Notification {
	return &Notification{
		Category: ROOM_JOINED,
		Alert: Alert{
			Body:      fmt.Sprintf("%s invited you to join the room \"%s\"", from, room),
			Key:       "room_invite_with_name_notification",
			Arguments: []string{from, room},
		},
		Arguments: map[string]interface{}{"id": id},
	}
}
