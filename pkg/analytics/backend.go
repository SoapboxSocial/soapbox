package analytics

import "database/sql"

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{db: db}
}

func (b *Backend) MarkNotificationRead(user int, uuid string) error {
	stmt, err := b.db.Prepare("UPDATE sent_notifications SET opened = NOW() WHERE user_id = $1 AND id = $2;")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user, uuid)
	return err
}
