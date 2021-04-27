package apple

import (
	"encoding/json"

	"github.com/sideshow/apns2"

	"github.com/soapboxsocial/soapbox/pkg/notifications"
)

type APNS struct {
	topic string

	client *apns2.Client
}

func NewAPNS(topic string, client *apns2.Client) *APNS {
	return &APNS{
		topic:  topic,
		client: client,
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

	// @todo handle response properly
	_, err = a.client.Push(payload)
	return err
}
