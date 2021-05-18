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
	stmt, err := b.db.Prepare("SELECT id, display_name, username FROM users WHERE id IN (SELECT recommendation FROM follow_recommendations WHERE id = $1);")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(user)
	if err != nil {
		return nil, err
	}

	result := make([]types.User, 0)

	for rows.Next() {
		user := types.User{}

		err := rows.Scan(&user.ID, &user.DisplayName, &user.Username)
		if err != nil {
			return nil, err // @todo
		}

		result = append(result, user)
	}

	return result, nil
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
	stmt, err := b.db.Prepare("INSERT INTO last_follow_recommended (user_id, last_recommended) VALUES ($1, $2);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user, time.Now())
	if err != nil {
		return err
	}

	return nil
}
