package worker

import "github.com/soapboxsocial/soapbox/pkg/notifications"

type Job struct {
	Targets      []notifications.Target
	Notification *notifications.PushNotification
}
