package stories

import "database/sql"

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{db}
}

func (b *Backend) GetStoriesForUser(user int) ([]*Story, error) {
	stmt, err := b.db.Prepare("SELECT id, expires_at, device_timestamp FROM stories WHERE user_id = $1 ORDER BY device_timestamp;")
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

		err := rows.Scan(&story.ID, &story.ExpiresAt, &story.DeviceTimestamp)
		if err != nil {
			return nil, err // @todo
		}

		result = append(result, story)
	}

	return result, nil
}

func (b *Backend) DeleteStory(story, user int) error {
	query := "DELETE FROM stories WHERE id = $1 AND user_id = $2;"

	stmt, err := b.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(story, user)
	return err
}
