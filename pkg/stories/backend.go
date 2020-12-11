package stories

import (
	"database/sql"
)

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{db}
}

// DeleteExpired deletes all stories where the expire_at time has passed and returns their IDs.
func (b *Backend) DeleteExpired(time int64) ([]string, error) {
	stmt, err := b.db.Prepare("DELETE FROM stories WHERE expires_at <= $1 RETURNING id;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(time)

	result := make([]string, 0)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var id string

		err := rows.Scan(&id)
		if err != nil {
			continue
		}

		result = append(result, id)
	}

	return result, nil
}

func (b *Backend) GetStoriesForUser(user int, time int64) ([]*Story, error) {
	stmt, err := b.db.Prepare("SELECT id, expires_at, device_timestamp FROM stories WHERE user_id = $1 AND expires_at >= $2 ORDER BY device_timestamp;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(user, time)
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

		story.Reactions = make([]Reaction, 0)

		result = append(result, story)
	}

	return result, nil
}

func (b *Backend) DeleteStory(story string, user int) error {
	query := "DELETE FROM stories WHERE id = $1 AND user_id = $2;"

	stmt, err := b.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(story, user)
	return err
}

func (b *Backend) AddStory(story string, user int, expires, timestamp int64) error {
	stmt, err := b.db.Prepare("INSERT INTO stories (id, user_id, expires_at, device_timestamp) VALUES ($1, $2, $3, $4);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(story, user, expires, timestamp)
	return err
}
