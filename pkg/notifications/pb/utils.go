package pb

import "github.com/soapboxsocial/soapbox/pkg/notifications"

func (x *Notification) ToPushNotification() *notifications.PushNotification {
	res := &notifications.PushNotification{
		Category:   notifications.NotificationCategory(x.Category),
		CollapseID: x.CollapseId,
		Arguments:  make(map[string]interface{}),
		Alert: notifications.Alert{
			Body:      x.Alert.Body,
			Key:       x.Alert.LocalizationKey,
			Arguments: x.Alert.LocalizationArguments,
		},
	}

	for key, arg := range x.Arguments {
		switch arg.Value.(type) {
		case *Notification_Argument_Int:
			res.Arguments[key] = arg.GetInt()
		case *Notification_Argument_Str:
			res.Arguments[key] = arg.GetStr()
		}
	}

	return res
}
