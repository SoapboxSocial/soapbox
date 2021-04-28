package notifications

import (
	"log"
	"time"

	"github.com/soapboxsocial/soapbox/pkg/devices"
)

type Service struct {
	apns    APNS
	limiter *Limiter
	devices *devices.Backend
	store   *Storage

	unregistered chan string
}

func NewService(apns APNS, limiter *Limiter, devices *devices.Backend, store *Storage) *Service {
	s := &Service{
		apns:         apns,
		limiter:      limiter,
		devices:      devices,
		store:        store,
		unregistered: make(chan string, 100),
	}

	go s.wipeDevices()

	return s
}

func (s *Service) Send(target Target, notification *PushNotification) {
	if !s.limiter.ShouldSendNotification(target, notification) {
		return
	}

	d, err := s.devices.GetDevicesForUser(target.ID)
	if err != nil {
		log.Printf("devicesBackend.GetDevicesForUser err: %v\n", err)
		return
	}

	for _, device := range d {
		go func(device string) {
			err = s.apns.Send(device, *notification)
			if err != nil {
				log.Printf("failed to send to target \"%s\" with error: %s\n", device, err)

				if err == ErrDeviceUnregistered {
					s.unregistered <- device
				}
			}
		}(device)
	}

	s.limiter.SentNotification(target, notification)

	store := getNotificationForStore(notification)
	if store == nil {
		return
	}

	err = s.store.Store(target.ID, store)
	if err != nil {
		log.Printf("notificationStorage.Store err: %v\n", err)
	}
}

func (s *Service) wipeDevices() {
	for device := range s.unregistered {
		log.Printf("removing device: %s", device)

		err := s.devices.RemoveDevice(device)
		if err != nil {
			log.Printf("failed to remove device err: %s", err)
		}
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
