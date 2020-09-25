package pubsub

import "github.com/go-redis/redis/v8"

type Topic string

type Event struct {

}

type Queue struct {
	buffer chan Event

	rdb *redis.Client
}

// NewQueue creates a new redis pubsub Queue.
func NewQueue(rdb *redis.Client) *Queue {
	return &Queue{
		buffer: make(chan Event, 100),
		rdb: rdb,
	}
}

// Publish an Event on a specific topic.
func (q *Queue) Publish(topic Topic, event Event)  {
	q.rdb.Publish(q.rdb.Context(), string(topic), event)
}

// Subscribe to a list of topics.
func (q *Queue) Subscribe(topics ...Topic) <-chan Event {
	t := make([]string, 0)
	for topic := range topics {
		t= append(t, string(topic))
	}

	pubsub := q.rdb.Subscribe(q.rdb.Context(), t...)
	go q.read(pubsub)

	return q.buffer
}

func (q *Queue) read(pubsub *redis.PubSub) {
	c := pubsub.Channel()
	for _ = range c {
		// @todo
	}
}
