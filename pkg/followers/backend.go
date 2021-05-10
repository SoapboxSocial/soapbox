package followers

import (
	"database/sql"

	"github.com/soapboxsocial/soapbox/pkg/users/types"
)

type FollowersBackend struct {
	db *sql.DB
}

func NewFollowersBackend(db *sql.DB) *FollowersBackend {
	return &FollowersBackend{
		db: db,
	}
}

func (fb *FollowersBackend) FollowUser(follower, user int) error {
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

func (fb *FollowersBackend) UnfollowUser(follower, user int) error {
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

func (fb *FollowersBackend) GetAllUsersFollowing(id, limit, offset int) ([]*types.User, error) {
	stmt, err := fb.db.Prepare("SELECT users.id, users.display_name, users.username, users.image FROM users INNER JOIN followers ON (users.id = followers.follower) WHERE followers.user_id = $1 ORDER BY users.id LIMIT $2 OFFSET $3;")
	if err != nil {
		return nil, err
	}

	return fb.executeUserQuery(stmt, id, limit, offset)
}

func (fb *FollowersBackend) GetAllUsersFollowedBy(id, limit, offset int) ([]*types.User, error) {
	stmt, err := fb.db.Prepare("SELECT users.id, users.display_name, users.username, users.image FROM users INNER JOIN followers ON (users.id = followers.user_id) WHERE followers.follower = $1 ORDER BY users.id LIMIT $2 OFFSET $3;")
	if err != nil {
		return nil, err
	}

	return fb.executeUserQuery(stmt, id, limit, offset)
}

func (fb *FollowersBackend) GetAllFollowerIDsFor(id int) ([]int, error) {
	stmt, err := fb.db.Prepare("SELECT follower FROM followers WHERE user_id = $1;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}

	result := make([]int, 0)

	for rows.Next() {
		var id int

		err := rows.Scan(&id)
		if err != nil {
			return nil, err // @todo
		}

		result = append(result, id)
	}

	return result, nil
}

func (fb *FollowersBackend) GetFriends(id int) ([]*types.User, error) {
	stmt, err := fb.db.Prepare("SELECT users.id, users.display_name, users.username, users.image FROM users WHERE id in (SELECT user_id AS user from followers WHERE follower = $1 INTERSECT SELECT follower as user FROM followers WHERE user_id = $1);")
	if err != nil {
		return nil, err
	}

	return fb.executeUserQuery(stmt, id)
}

func (fb *FollowersBackend) executeUserQuery(stmt *sql.Stmt, args ...interface{}) ([]*types.User, error) {
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	result := make([]*types.User, 0)

	for rows.Next() {
		user := &types.User{}

		err := rows.Scan(&user.ID, &user.DisplayName, &user.Username, &user.Image)
		if err != nil {
			return nil, err // @todo
		}

		result = append(result, user)
	}

	return result, nil
}
