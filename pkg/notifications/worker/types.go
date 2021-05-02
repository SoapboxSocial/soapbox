package worker

import "github.com/soapboxsocial/soapbox/pkg/notifications"

type Job struct {
	Origin       int
	Targets      []notifications.Target
	Notification *notifications.PushNotification
}
