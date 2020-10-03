package notifications

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

type Storage struct {
	rdb *redis.Client
}

func NewStorage(rdb *redis.Client) *Storage {
	return &Storage{
		rdb: rdb,
	}
}

func (s *Storage) Store(user int, notification *Notification) error {
	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	key := notificationListKey(user)
	err = s.rdb.LPush(s.rdb.Context(), key, string(data)).Err()
	if err != nil {
		return err
	}

	return s.rdb.LTrim(s.rdb.Context(), key, 0, 9).Err()
}

func (s *Storage) GetNotifications(user int) ([]*Notification, error) {
	data, err := s.rdb.LRange(s.rdb.Context(), notificationListKey(user), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	notifications := make([]*Notification, 0)
	for _, item := range data {
		n := &Notification{}
		err := json.Unmarshal([]byte(item), n)
		if err != nil {
			log.Printf("failed to unmarshal notification err: %v\n", err)
			continue
		}

		notifications = append(notifications, n)
	}

	return notifications, nil
}

func notificationListKey(user int) string {
	return fmt.Sprintf("notifications_%d", user)
}
