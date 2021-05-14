package follows

import "database/sql"

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{db: db}
}

func (b *Backend) AddRecommendations(user int, recommendations []int) error {
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
