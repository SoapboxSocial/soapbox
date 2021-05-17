package follows

import (
	"database/sql"
	"time"

	"github.com/soapboxsocial/soapbox/pkg/users/types"
)

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{db: db}
}

func (b *Backend) RecommendationsFor(user int) ([]types.User, error) {
	return nil, nil
}

func (b *Backend) AddRecommendationsFor(user int, recommendations []int) error {
	tx, err := b.db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO follow_recommendations (user_id, recommendation) VALUES ($1, $2)")
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	for _, id := range recommendations {
		_, err = stmt.Exec(user, id)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}

func (b *Backend) LastUpdatedFor(user int) (*time.Time, error) {
	stmt, err := b.db.Prepare("SELECT last_recommended FROM last_follow_recommended WHERE user_id = $1;")
	if err != nil {
		return nil, err
	}

	timestamp := &time.Time{}
	err = stmt.QueryRow(user).Scan(timestamp)
	if err != nil {
		return nil, err
	}

	return timestamp, nil
}

func (b *Backend) SetLastUpdatedFor(user int) error {
	return nil
}
