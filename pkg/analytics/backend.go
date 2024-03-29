package analytics

import (
	"database/sql"
)

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{db: db}
}

func (b *Backend) AddSentNotification(user int, notification Notification) error {
	stmt, err := b.db.Prepare("INSERT INTO notification_analytics (id, target, origin, category, sent, room) VALUES($1, $2, $3, $4, NOW(), $5);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(notification.ID, user, notification.Origin, notification.Category, notification.Room)
	return err
}

func (b *Backend) MarkNotificationRead(user int, uuid string) error {
	stmt, err := b.db.Prepare("UPDATE notification_analytics SET opened = NOW() WHERE target = $1 AND id = $2;")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user, uuid)
	return err
}
