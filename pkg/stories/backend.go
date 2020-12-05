package stories

import "database/sql"

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{db}
}

func (b *Backend) GetStoriesForUser(user int) ([]*Story, error) {
	stmt, err := b.db.Prepare("SELECT id, expires_at FROM stories WHERE user_id = $1 ORDER BY expires_at ASC;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(user)
	if err != nil {
		return nil, err
	}

	result := make([]*Story, 0)

	for rows.Next() {
		story := &Story{}

		err := rows.Scan(&story.ID, &story.ExpiresAt)
		if err != nil {
			return nil, err // @todo
		}

		result = append(result, story)
	}

	return result, nil
}
