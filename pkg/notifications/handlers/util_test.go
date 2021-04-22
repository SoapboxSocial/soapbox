package handlers_test

import (
	"encoding/json"

	"github.com/soapboxsocial/soapbox/pkg/pubsub"
)

func getRawEvent(event *pubsub.Event) (*pubsub.Event, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	evt := &pubsub.Event{}
	err = json.Unmarshal(data, evt)
	if err != nil {
		return nil, err
	}

	return evt, nil
}
