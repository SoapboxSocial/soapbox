package worker

import (
	"log"
	"time"

	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
)

type Config struct {
	APNS    notifications.APNS
	Limiter *notifications.Limiter
	Devices *devices.Backend
	Store   *notifications.Storage
}

type Worker struct {
	jobs    chan Job
	workers chan<- chan Job
	quit    chan bool

	unregistered chan string
	config       *Config
}

func NewWorker(pool chan<- chan Job, config *Config) *Worker {
	return &Worker{
		workers:      pool,
		jobs:         make(chan Job),
		quit:         make(chan bool),
		unregistered: make(chan string, 100),
		config:       config,
	}
}

func (w *Worker) Start() {

	go w.wipeDevices()

	go func() {
		for {
			w.workers <- w.jobs

			select {
			case job := <-w.jobs:
				// Receive a work request.
				w.handle(job)
			case <-w.quit:
				// We have been asked to stop.
				close(w.unregistered)
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func (w *Worker) handle(job Job) {
	if !w.config.Limiter.ShouldSendNotification(job.Target, job.Notification) {
		return
	}

	d, err := w.config.Devices.GetDevicesForUser(job.Target.ID)
	if err != nil {
		log.Printf("devicesBackend.GetDevicesForUser err: %v\n", err)
		return
	}

	for _, device := range d {
		go func(device string) {
			err = w.config.APNS.Send(device, *job.Notification)
			if err != nil {
				log.Printf("failed to send to target \"%s\" with error: %s\n", device, err)

				if err == notifications.ErrDeviceUnregistered {
					w.unregistered <- device
				}
			}
		}(device)
	}

	w.config.Limiter.SentNotification(job.Target, job.Notification)

	store := getNotificationForStore(job.Notification)
	if store == nil {
		return
	}

	err = w.config.Store.Store(job.Target.ID, store)
	if err != nil {
		log.Printf("notificationStorage.Store err: %v\n", err)
	}
}

func (w *Worker) wipeDevices() {
	for device := range w.unregistered {
		log.Printf("removing device: %s", device)

		err := w.config.Devices.RemoveDevice(device)
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
