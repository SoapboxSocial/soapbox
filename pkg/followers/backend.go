package followers

import (
	"database/sql"

	"github.com/ephemeral-networks/voicely/pkg/users"
)

type FollowersBackend struct {
	db *sql.DB
}

func NewFollowersBackend(db *sql.DB) *FollowersBackend {
	return &FollowersBackend{
		db: db,
	}
}

func (fb *FollowersBackend) FollowUser(follower int, user int) error {
	stmt, err := fb.db.Prepare("INSERT INTO followers (follower, user_id) VALUES ($1, $2);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(follower, user)
	if err != nil {
		return err
	}

	return nil
}

func (fb *FollowersBackend) UnfollowUser(follower int, user int) error {
	stmt, err := fb.db.Prepare("DELETE FROM followers WHERE follower = $1 AND user_id = $2;")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(follower, user)
	if err != nil {
		return err
	}

	return nil
}

func (fb *FollowersBackend) GetAllUsersFollowing(id int) ([]*users.User, error) {
	stmt, err := fb.db.Prepare("SELECT users.id, users.display_name, users.username  FROM users INNER JOIN followers ON (users.id = followers.follower) WHERE followers.user_id = $1;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}

	result := make([]*users.User, 0)

	for rows.Next() {
		user := &users.User{}

		err := rows.Scan(user.ID, user.DisplayName, user.Username)
		if err != nil {
			return nil, err // @todo
		}

		result = append(result, user)
	}

	return result, nil
}

func (fb *FollowersBackend) GetAllUsersFollowedBy(id int) ([]*users.User, error) {
	stmt, err := fb.db.Prepare("SELECT users.id, users.display_name, users.username  FROM users INNER JOIN followers ON (users.id = followers.follower) WHERE followers.follower = $1;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}

	result := make([]*users.User, 0)

	for rows.Next() {
		user := &users.User{}

		err := rows.Scan(&user.ID, &user.DisplayName, &user.Username)
		if err != nil {
			return nil, err // @todo
		}

		result = append(result, user)
	}

	return result, nil
}