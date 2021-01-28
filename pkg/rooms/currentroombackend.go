package rooms

import (
	"strconv"

	"github.com/go-redis/redis/v8"
)

const hashName = "current_room"

type CurrentRoomBackend struct {
	db *redis.Client
}

func NewCurrentRoomBackend(db *redis.Client) *CurrentRoomBackend {
	return &CurrentRoomBackend{
		db: db,
	}
}

func (b *CurrentRoomBackend) GetCurrentRoomForUser(id int) (string, error) {
	val, err := b.db.HGet(b.db.Context(), hashName, strconv.Itoa(id)).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}

func (b *CurrentRoomBackend) SetCurrentRoomForUser(user int, room string) error {
	_, err := b.db.HSet(b.db.Context(), hashName, strconv.Itoa(user), room).Result()
	return err
}

func (b *CurrentRoomBackend) RemoveCurrentRoomForUser(user int) {
	b.db.HDel(b.db.Context(), hashName, strconv.Itoa(user))
}
