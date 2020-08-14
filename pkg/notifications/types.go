package notifications

type NotificationCategory string

const (
	NEW_ROOM     NotificationCategory = "NEW_ROOM"
	NEW_FOLLOWER                      = "NEW_FOLLOWER"
)

type Alert struct {
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
