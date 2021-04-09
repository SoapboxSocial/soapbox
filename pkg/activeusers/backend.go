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
	stmt, err := b.db.Prepare("SELECT update_user_active_times($1, $2);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user,time)
	return err}
