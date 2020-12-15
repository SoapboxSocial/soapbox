package stories

import (
	"database/sql"
	"errors"
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

		reactions, err := b.GetReactions(story.ID)
		if err != nil {
			continue
		}

		story.Reactions = reactions

		result = append(result, story)
	}

	return result, nil
}

func (b *Backend) GetStoriesForFollower(user int, time int64) (map[int][]Story, error) {
	stmt, err := b.db.Prepare("SELECT stories.user_id, stories.id, stories.expires_at, stories.device_timestamp FROM stories INNER JOIN followers ON (stories.user_id = followers.user_id) WHERE followers.follower = $1 AND stories.expires_at >= $2;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(user, time)
	if err != nil {
		return nil, err
	}

	result := make(map[int][]Story, 0)

	for rows.Next() {
		story := Story{}

		var user int
		err := rows.Scan(&user, &story.ID, &story.ExpiresAt, &story.DeviceTimestamp)
		if err != nil {
			return nil, err // @todo
		}

		reactions, err := b.GetReactions(story.ID)
		if err != nil {
			continue
		}

		story.Reactions = reactions

		_, ok := result[user]
		if !ok {
			result[user] = make([]Story, 0)
		}

		result[user] = append(result[user], story)
	}

	return result, nil
}

func (b *Backend) DeleteStory(story string, user int) error {
	query := "DELETE FROM stories WHERE id = $1 AND user_id = $2;"

	stmt, err := b.db.Prepare(query)
	if err != nil {
		return err
	}

	res, err := stmt.Exec(story, user)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return errors.New("no story deleted")
	}

	return nil
}

func (b *Backend) AddStory(story string, user int, expires, timestamp int64) error {
	stmt, err := b.db.Prepare("INSERT INTO stories (id, user_id, expires_at, device_timestamp) VALUES ($1, $2, $3, $4);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(story, user, expires, timestamp)
	return err
}

func (b *Backend) ReactToStory(story, reaction string, user int) error {
	stmt, err := b.db.Prepare("INSERT INTO story_reactions (story_id, user_id, reaction) VALUES ($1, $2, $3);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(story, user, reaction)
	return err
}

func (b *Backend) GetReactions(story string) ([]Reaction, error) {
	reactions := make([]Reaction, 0)
	stmt, err := b.db.Prepare("SELECT reaction, COUNT(*) FROM story_reactions WHERE story_id = $1 GROUP BY reaction;")
	if err != nil {
		return reactions, err
	}

	rows, err := stmt.Query(story)
	if err != nil {
		return reactions, err
	}

	for rows.Next() {
		reaction := Reaction{}

		err := rows.Scan(&reaction.Emoji, &reaction.Count)
		if err != nil {
			return nil, err // @todo
		}

		reactions = append(reactions, reaction)
	}

	return reactions, nil
}
