package notifications

import "github.com/sideshow/apns2"

type Service struct {
	topic string

	client *apns2.Client
}

func NewService(topic string, client *apns2.Client) {

}