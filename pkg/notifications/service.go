package notifications

import (
	"encoding/json"
	"fmt"

	"github.com/sideshow/apns2"
)

type Service struct {
	topic string

	client *apns2.Client
}

func NewService(topic string, client *apns2.Client) *Service {
	return &Service{
		topic:  topic,
		client: client,
	}
}

func (s *Service) Send(target string, notification Notification) error {
	data, err := json.Marshal(map[string]interface{}{"aps": notification})
	if err != nil {
		return err
	}

	fmt.Println(string(data))

	payload := &apns2.Notification{}
	payload.DeviceToken = target
	payload.Topic = s.topic
	payload.Payload = data

	// @todo handle response properly
	_, err = s.client.Push(payload)
	return err
}
