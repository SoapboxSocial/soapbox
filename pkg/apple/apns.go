package apple

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sideshow/apns2"

	"github.com/soapboxsocial/soapbox/pkg/notifications"
)

type APNS struct {
	topic string

	client *apns2.Client

	// maxConcurrentPushes limits the amount of notification pushes
	// this is required for apple, not sure if it should be here tho.
	maxConcurrentPushes chan struct{}
}

func NewAPNS(topic string, client *apns2.Client) *APNS {
	return &APNS{
		topic:               topic,
		client:              client,
		maxConcurrentPushes: make(chan struct{}, 100),
	}
}

func (a *APNS) Send(target string, notification notifications.PushNotification) error {
	data, err := json.Marshal(map[string]interface{}{"aps": notification})
	if err != nil {
		return err
	}

	payload := &apns2.Notification{
		DeviceToken: target,
		Topic:       a.topic,
		Payload:     data,
		CollapseID:  notification.CollapseID,
	}

	a.maxConcurrentPushes <- struct{}{}
	resp, err := a.client.Push(payload)
	<-a.maxConcurrentPushes
	if err != nil {
		return err
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return notifications.ErrRetryRequired
	}

	if resp.Reason == apns2.ReasonUnregistered {
		return notifications.ErrDeviceUnregistered
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send code: %d reason: %s", resp.StatusCode, resp.Reason)
	}

	return nil
}
