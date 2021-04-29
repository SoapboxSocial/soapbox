package worker

import "github.com/soapboxsocial/soapbox/pkg/notifications"

type Job struct {
	Target       notifications.Target
	Notification *notifications.PushNotification
}
