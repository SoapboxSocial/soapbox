package handlers

import (
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
)

// Handler handles a specific type of notification
type Handler interface {

	// Type returns the event handled to build a notification
	Type() pubsub.EventType

	// Targets returns the notification receivers
	Targets(event *pubsub.Event) ([]notifications.Target, error)

	// Build builds the notification
	Build(event *pubsub.Event) (*notifications.PushNotification, error)
}
