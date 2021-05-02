package worker

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/soapboxsocial/soapbox/pkg/analytics"
	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
)

type Config struct {
	APNS      notifications.APNS
	Limiter   *notifications.Limiter
	Devices   *devices.Backend
	Store     *notifications.Storage
	Analytics *analytics.Backend
}

type Worker struct {
	jobs    chan Job
	workers chan<- chan Job
	quit    chan bool

	unregistered chan string
	config       *Config

	maxRetries int
}

func NewWorker(pool chan<- chan Job, config *Config) *Worker {
	return &Worker{
		workers:      pool,
		jobs:         make(chan Job),
		quit:         make(chan bool),
		unregistered: make(chan string, 100),
		config:       config,
		maxRetries:   5,
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
	ids := make([]int, 0)
	targets := make([]notifications.Target, 0)

	for _, target := range job.Targets {
		if !w.config.Limiter.ShouldSendNotification(target, job.Notification) {
			continue
		}

		ids = append(ids, target.ID)
		targets = append(targets, target)
	}

	if len(ids) == 0 {
		return
	}

	d, err := w.config.Devices.GetDevicesForUsers(ids)
	if err != nil {
		log.Printf("devicesBackend.GetDevicesForUsers err: %v\n", err)
		return
	}

	log.Printf("pushing %s to %d targets", job.Notification.Category, len(targets))

	notification := *job.Notification
	notification.UUID = uuid.NewString()

	for i := 0; i < w.maxRetries; i++ {
		d = w.sendNotifications(d, notification)
		if len(d) == 0 {
			break
		}
	}

	for _, target := range targets {
		an := job.Notification.AnalyticsNotification()
		if job.Origin != 0 {
			an.Origin = &job.Origin
		}

		err := w.config.Analytics.AddSentNotification(target.ID, an)
		if err != nil {
			log.Printf("analytics.AddSentNotification err: %s\n", err)
		}

		w.config.Limiter.SentNotification(target, job.Notification)

		store := getNotificationForStore(job.Notification)
		if store == nil {
			return
		}

		err = w.config.Store.Store(target.ID, store)
		if err != nil {
			log.Printf("notificationStorage.Store err: %v\n", err)
		}
	}
}

// @TODO THIS SHOULD PROBABLY BE MOVED INTO APNS, especially once we add iOS
func (w *Worker) sendNotifications(devices []string, notification notifications.PushNotification) []string {
	var wg sync.WaitGroup

	retry := make([]string, 0)

	for _, device := range devices {
		wg.Add(1)
		go func(device string) {
			err := w.config.APNS.Send(device, notification)
			if err != nil {
				switch err {
				case notifications.ErrDeviceUnregistered:
					w.unregistered <- device
				case notifications.ErrRetryRequired:
					retry = append(retry, device)
				}

				log.Printf("failed to send to target \"%s\" with error: %s\n", device, err)
			}

			wg.Done()
		}(device)
	}

	wg.Wait()

	return retry
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
