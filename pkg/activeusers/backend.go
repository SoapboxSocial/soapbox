package activeusers

import (
	"database/sql"
	"time"
)

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{
		db: db,
	}
}

func (b *Backend) SetLastActiveTime(user int, time time.Time) error {
	return nil
}
