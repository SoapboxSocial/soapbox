package notifications

type APNS interface {
	Send(target string, notification PushNotification) error
}
