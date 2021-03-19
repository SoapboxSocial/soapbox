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

func (b *UserRoomLogBackend) Store(user int, room string, joined, left time.Time) error {
	return nil
}
