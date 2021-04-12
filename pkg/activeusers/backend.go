package activeusers

import (
	"database/sql"
	"time"
)

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{
		db: db,
	}
}

func (b *Backend) SetLastActiveTime(user int, time time.Time) error {
	stmt, err := b.db.Prepare("SELECT update_user_active_times($1, $2);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user, time)
	return err
}

func (b *Backend) GetActiveUsersForFollower(user int) ([]ActiveUser, error) {
	query := `SELECT users.id, users.display_name, users.username, users.image, active.room FROM users
		INNER JOIN (SELECT * FROM current_rooms UNION (SELECT user_id, '' FROM user_active_times WHERE last_active > (NOW() - INTERVAL '30 MINUTE'))) active ON users.id = active.user_id 
		WHERE active.user_id IN (SELECT user_id AS user from followers WHERE follower = $1 INTERSECT SELECT follower as user FROM followers WHERE user_id = $1);`

	stmt, err := b.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(user)
	if err != nil {
		return nil, err
	}

	result := make([]ActiveUser, 0)

	for rows.Next() {
		user := ActiveUser{}
		var room string

		err := rows.Scan(&user.ID, &user.DisplayName, &user.Username, &user.Image, &room)
		if err != nil {
			continue
		}

		if room != "" {
			user.Room = &room
		}

		result = append(result, user)
	}

	return result, nil
}
