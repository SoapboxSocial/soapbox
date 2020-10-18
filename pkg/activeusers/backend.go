package activeusers

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
)

type User struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Image       string `json:"image"`
	CurrentRoom int    `json:"current_room"`
}

type Backend struct {
	redis *redis.Client
	db    *sql.DB
}

func (b *Backend) FetchActiveUsersFollowedBy(id int) ([]*User, error) {
	keys, err := b.redis.HKeys(b.redis.Context(), "current_room").Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, nil
	}

	raw := "SELECT users.id, users.display_name, users.username, users.image FROM users INNER JOIN followers ON (users.id = followers.user_id) WHERE followers.follower = %d AND followers.user_id IN (%s);"
	query := fmt.Sprintf(raw, id, strings.Join(keys, ", "))

	rows, err := b.db.Query(query)
	if err != nil {
		return nil, err
	}

	result := make([]*User, 0)

	for rows.Next() {
		user := &User{}

		err := rows.Scan(&user.ID, &user.DisplayName, &user.Username, &user.Image)
		if err != nil {
			return nil, err // @todo
		}

		val, err := b.redis.HGet(b.redis.Context(), "current_room", strconv.Itoa(user.ID)).Result()
		if err != nil {
			continue
		}

		id, err = strconv.Atoi(val)
		if err != nil {
			continue
		}

		user.CurrentRoom = id

		result = append(result, user)
	}

	return result, nil
}
