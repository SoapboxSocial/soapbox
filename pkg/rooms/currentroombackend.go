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

func (b *CurrentRoomBackend) GetCurrentRoomForUser(id int) (int, error) {
	val, err := b.db.HGet(b.db.Context(), hashName, string(id)).Result()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(val)
}

func (b *CurrentRoomBackend) SetCurrentRoomForUser(user, room int) error {
	b.db.HSet(b.db.Context(), hashName, string(user), strconv.Itoa(room))
	return nil
}

func (b *CurrentRoomBackend) RemoveCurrentRoomForUser(user int) {
	b.db.HDel(b.db.Context(), hashName, string(user))
}