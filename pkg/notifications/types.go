package notifications

type NotificationCategory string

const (
	NEW_ROOM     NotificationCategory = "NEW_ROOM"
	NEW_FOLLOWER NotificationCategory = "NEW_FOLLOWER"
	ROOM_INVITE  NotificationCategory = "ROOM_INVITE"
	ROOM_JOINED  NotificationCategory = "ROOM_JOINED"
	WELCOME_ROOM NotificationCategory = "WELCOME_ROOM"
)

type Frequency int

const (
	FrequencyOff = iota
	Infrequent
	Normal
	Frequent
)

// Target represents the notification target and their settings.
type Target struct {
	ID            int       `json:"-"`
	RoomFrequency Frequency `json:"room_frequency"`
	Follows       bool      `json:"follows"`
}

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

func NewRoomNotification(id, creator string, creatorID int) *PushNotification {
	return &PushNotification{
		Category: NEW_ROOM,
		Alert: Alert{
			Key:       "new_room_notification",
			Arguments: []string{creator},
		},
		Arguments:  map[string]interface{}{"id": id, "creator": creatorID},
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

func NewRoomInviteNotification(id, from string) *PushNotification {
	return &PushNotification{
		Category: ROOM_INVITE,
		Alert: Alert{
			Key:       "room_invite_notification",
			Arguments: []string{from},
		},
		Arguments:  map[string]interface{}{"id": id},
		CollapseID: id,
	}
}

func NewRoomInviteNotificationWithName(id, from, room string) *PushNotification {
	return &PushNotification{
		Category: ROOM_INVITE,
		Alert: Alert{
			Key:       "room_invite_with_name_notification",
			Arguments: []string{from, room},
		},
		Arguments:  map[string]interface{}{"id": id},
		CollapseID: id,
	}
}
