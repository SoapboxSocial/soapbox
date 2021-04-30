package pb

import "github.com/soapboxsocial/soapbox/pkg/notifications"

func (x *Notification) ToPushNotification() *notifications.PushNotification {
	res := &notifications.PushNotification{
		Category:   notifications.NotificationCategory(x.Category),
		CollapseID: x.CollapseId,
		Alert: notifications.Alert{
			Body:      x.Alert.Body,
			Key:       x.Alert.LocalizationKey,
			Arguments: x.Alert.LocalizationArguments,
		},
	}

	return res
}
