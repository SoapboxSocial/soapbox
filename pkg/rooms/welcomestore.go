package rooms

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const timeout = 30 * time.Minute

type WelcomeStore struct {
	rdb *redis.Client
}

func NewWelcomeStore(rdb *redis.Client) *WelcomeStore {
	return &WelcomeStore{rdb: rdb}
}

func (w *WelcomeStore) StoreWelcomeRoomID(room string, user int64) error {
	_, err := w.rdb.Set(w.rdb.Context(), key(room), strconv.Itoa(int(user)), timeout).Result()
	return err
}

func (w *WelcomeStore) GetUserIDForWelcomeRoom(room string) (int, error) {
	str, err := w.rdb.Get(w.rdb.Context(), key(room)).Result()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(str)
}

func (w *WelcomeStore) DeleteWelcomeRoom(room string) error {
	_, err := w.rdb.Del(w.rdb.Context(), key(room)).Result()
	return err
}

func key(room string) string {
	return fmt.Sprintf("welcome_room_%s", room)
}
