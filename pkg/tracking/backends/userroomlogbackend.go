package backends

import (
	"database/sql"
	"time"
)

type UserRoomLogBackend struct {
	db *sql.DB
}

func NewUserRoomLogBackend(db *sql.DB) *UserRoomLogBackend {
	return &UserRoomLogBackend{db: db}
}

func (b *UserRoomLogBackend) Store(user int, room, visibility string, joined, left time.Time) error {
	stmt, err := b.db.Prepare("INSERT INTO user_room_logs (user_id, room, join_time, left_time, visibility) VALUES ($1, $2, $3, $4, $5);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user, room, joined, left, visibility)
	return err
}
