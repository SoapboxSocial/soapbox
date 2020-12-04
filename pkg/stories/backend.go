package stories

import "database/sql"

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{db}
}

func (b *Backend) GetStoriesForUser(user int) ([]string, error) {
	stmt, err := b.db.Prepare("SELECT id FROM stores WHERE user_id $1 ORDER BY timestamp;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(user)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for rows.Next() {
		var id string

		err := rows.Scan(&id)
		if err != nil {
			return nil, err // @todo
		}

		result = append(result, id)
	}

	return result, nil
}