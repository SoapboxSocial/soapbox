package indexer

import (
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

const queueName = "indexer_queue"

type EventType int

const (
	EventTypeUserUpdate EventType = iota
)

type Event struct {
	Type   EventType              `json:"type"`
	Params map[string]interface{} `json:"params"`
}

type Queue struct {
	db *redis.Client
}

func NewIndexerQueue(db *redis.Client) *Queue {
	return &Queue{
		db: db,
	}
}

func (q *Queue) Len() int {
	val, err := q.db.LLen(q.db.Context(), queueName).Result()
	if err != nil {
		return 0
	}

	return int(val)
}

func (q *Queue) Push(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	q.db.LPush(q.db.Context(), queueName, string(data))
}

func (q *Queue) Pop() (*Event, error) {
	result, err := q.db.LPop(q.db.Context(), queueName).Result()
	if err != nil {
		return nil, err
	}

	event := &Event{}
	err = json.Unmarshal([]byte(result), event)
	if err != nil {
		return nil, err
	}

	return event, nil
}
