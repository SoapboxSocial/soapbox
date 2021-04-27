package notifications

import (
	"log"
	"time"

	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
)

type Service struct {
	apns    *APNS
	limiter *Limiter
	devices *devices.Backend
	store   *Storage
}

func NewService(apns *APNS, limiter *Limiter, devices *devices.Backend, store *Storage) *Service {
	return &Service{
		apns:    apns,
		limiter: limiter,
		devices: devices,
		store:   store,
	}
}

func (s *Service) Send(target Target, event *pubsub.Event, notification *PushNotification) {
	if !s.limiter.ShouldSendNotification(target, event) {
		return
	}

	d, err := s.devices.GetDevicesForUser(target.ID)
	if err != nil {
		log.Printf("devicesBackend.GetDevicesForUser err: %v\n", err)
	}

	for _, device := range d {
		err = s.apns.Send(device, *notification)
		if err != nil {
			log.Printf("failed to send to target \"%s\" with error: %s\n", device, err.Error())
		}
	}

	s.limiter.SentNotification(target, event)

	store := getNotificationForStore(notification)
	if store == nil {
		return
	}

	err = s.store.Store(target.ID, store)
	if err != nil {
		log.Printf("notificationStorage.Store err: %v\n", err)
	}
}

// @TODO MOVE TO HANDLER
func getNotificationForStore(notification *PushNotification) *Notification {
	switch notification.Category {
	case NEW_FOLLOWER:
		return &Notification{
			Timestamp: time.Now().Unix(),
			From:      notification.Arguments["id"].(int),
			Category:  notification.Category,
		}
	case WELCOME_ROOM:
		return &Notification{
			Timestamp: time.Now().Unix(),
			From:      notification.Arguments["from"].(int),
			Category:  notification.Category,
			Arguments: map[string]interface{}{"room": notification.Arguments["id"]},
		}
	default:
		return nil
	}
}
