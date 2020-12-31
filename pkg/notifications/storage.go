package notifications

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

const placeholder = "val"

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

	s.setHasNewNotifications(user)

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

func (s *Storage) MarkNotificationsViewed(user int) {
	s.rdb.Del(s.rdb.Context(), hasNewNotificationsKey(user))
}

func (s *Storage) HasNewNotifications(user int) bool {
	res, err := s.rdb.Get(s.rdb.Context(), hasNewNotificationsKey(user)).Result()
	if err != nil {
		return false
	}

	return res == placeholder
}

func (s *Storage) setHasNewNotifications(user int) {
	s.rdb.Set(s.rdb.Context(), hasNewNotificationsKey(user), placeholder, 0)
}

func hasNewNotificationsKey(user int) string {
	return fmt.Sprintf("has_new_notifications_%d", user)
}

func notificationListKey(user int) string {
	return fmt.Sprintf("notifications_%d", user)
}
