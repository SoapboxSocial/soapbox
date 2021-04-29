package worker

import (
	"log"
	"time"

	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
)

type Config struct {
	apns    notifications.APNS
	limiter *notifications.Limiter
	devices *devices.Backend
	store   *notifications.Storage
}

type Worker struct {
	Work        chan Job
	WorkerQueue chan chan Job
	QuitChan    chan bool

	unregistered chan string
	config       *Config
}

func (w *Worker) Start() {

	go w.wipeDevices()

	go func() {
		for {
			w.WorkerQueue <- w.Work

			select {
			case job := <-w.Work:
				// Receive a work request.
				w.handle(job)
			case <-w.QuitChan:
				// We have been asked to stop.
				close(w.unregistered)
				return
			}
		}
	}()
}

func (w *Worker) handle(job Job) {
	if !w.config.limiter.ShouldSendNotification(job.Target, job.Notification) {
		return
	}

	d, err := w.config.devices.GetDevicesForUser(job.Target.ID)
	if err != nil {
		log.Printf("devicesBackend.GetDevicesForUser err: %v\n", err)
		return
	}

	for _, device := range d {
		go func(device string) {
			err = w.config.apns.Send(device, *job.Notification)
			if err != nil {
				log.Printf("failed to send to target \"%s\" with error: %s\n", device, err)

				if err == notifications.ErrDeviceUnregistered {
					w.unregistered <- device
				}
			}
		}(device)
	}

	w.config.limiter.SentNotification(job.Target, job.Notification)

	store := getNotificationForStore(job.Notification)
	if store == nil {
		return
	}

	err = w.config.store.Store(job.Target.ID, store)
	if err != nil {
		log.Printf("notificationStorage.Store err: %v\n", err)
	}
}

func (w *Worker) wipeDevices() {
	for device := range w.unregistered {
		log.Printf("removing device: %s", device)

		err := w.config.devices.RemoveDevice(device)
		if err != nil {
			log.Printf("failed to remove device err: %s", err)
		}
	}
}

func getNotificationForStore(notification *notifications.PushNotification) *notifications.Notification {
	switch notification.Category {
	case notifications.NEW_FOLLOWER:
		return &notifications.Notification{
			Timestamp: time.Now().Unix(),
			From:      notification.Arguments["id"].(int),
			Category:  notification.Category,
		}
	case notifications.WELCOME_ROOM:
		return &notifications.Notification{
			Timestamp: time.Now().Unix(),
			From:      notification.Arguments["from"].(int),
			Category:  notification.Category,
			Arguments: map[string]interface{}{"room": notification.Arguments["id"]},
		}
	default:
		return nil
	}
}
