package pubsub

import (
	"encoding/json"
	"log"

	"github.com/go-redis/redis/v8"
)

type Topic string

const (
	RoomTopic  Topic = "room"
	UserTopic  Topic = "user"
	StoryTopic Topic = "story"
)

type Queue struct {
	buffer chan *Event

	rdb *redis.Client
}

// NewQueue creates a new redis pubsub Queue.
func NewQueue(rdb *redis.Client) *Queue {
	return &Queue{
		buffer: make(chan *Event, 100),
		rdb:    rdb,
	}
}

// Publish an Event on a specific topic.
func (q *Queue) Publish(topic Topic, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	q.rdb.Publish(q.rdb.Context(), string(topic), data)
	return nil
}

// Subscribe to a list of topics.
func (q *Queue) Subscribe(topics ...Topic) <-chan *Event {
	t := make([]string, 0)
	for _, topic := range topics {
		t = append(t, string(topic))
	}

	pubsub := q.rdb.Subscribe(q.rdb.Context(), t...)
	go q.read(pubsub)

	return q.buffer
}

func (q *Queue) read(pubsub *redis.PubSub) {
	c := pubsub.Channel()
	for msg := range c {
		event := &Event{}
		err := json.Unmarshal([]byte(msg.Payload), event)
		if err != nil {
			log.Printf("failed to decode event err: %v event: %s", err, msg.Payload)
			continue
		}

		q.buffer <- event
	}
}
