package queue

import "github.com/go-redis/redis/v8"

type Event struct {

}

type Queue struct {
	rdb redis.Client
}

func (q *Queue) Publish(topic string, event Event)  {
	q.rdb.Publish(q.rdb.Context(), topic, event)
}
