package notifications

import (
	"encoding/json"
	"errors"

	"github.com/go-redis/redis/v8"
)

const queueName = "notification_queue"

type EventType int

const (
	EventTypeRoomCreation EventType = iota
)

type Event struct {
	Type    EventType              `json:"type"`
	Creator int                    `json:"creator"`
	Params  map[string]interface{} `json:"params"`
}

type Queue struct {
	db *redis.Client
}

func NewNotificationQueue(db *redis.Client) *Queue {
	return &Queue{
		db: db,
	}
}

func (q *Queue) Push(event Event) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	q.db.LPush(q.db.Context(), queueName, string(data))
}

func (q *Queue) Pop() (*Event, error) {
	if q.db.LLen(q.db.Context(), queueName).Val() == 0 {
		return nil, errors.New("no data")
	}

	result := q.db.LPop(q.db.Context(), queueName)

	event := &Event{}
	err := json.Unmarshal([]byte(result.Val()), event)
	if err != nil {
		return nil, err
	}

	return event, nil
}
