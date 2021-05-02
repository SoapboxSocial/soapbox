package notifications

import (
	"database/sql"
	"fmt"
	"strconv"
)

type Settings struct {
	db *sql.DB
}

func NewSettings(db *sql.DB) *Settings {
	return &Settings{db: db}
}

func (s *Settings) GetSettingsFor(user int) (*Target, error) {
	stmt, err := s.db.Prepare("SELECT user_id, room_frequency, follows FROM notification_settings WHERE user_id = $1;")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(user)

	target := &Target{}
	err = row.Scan(&target.ID, &target.RoomFrequency, &target.Follows)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func (s *Settings) GetSettingsFollowingUser(user int) ([]Target, error) {
	stmt, err := s.db.Prepare("SELECT notification_settings.user_id, notification_settings.room_frequency, notification_settings.follows FROM notification_settings INNER JOIN followers ON (notification_settings.user_id = followers.follower) WHERE followers.user_id = $1")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(user)
	if err != nil {
		return nil, err
	}

	targets := make([]Target, 0)
	for rows.Next() {
		target := Target{}
		err = rows.Scan(&target.ID, &target.RoomFrequency, &target.Follows)
		if err != nil {
			continue
		}

		targets = append(targets, target)
	}

	return targets, nil
}

func (s *Settings) GetSettingsForUsers(users []int64) ([]Target, error) {
	query := fmt.Sprintf(
		"SELECT notification_settings.user_id, notification_settings.room_frequency, notification_settings.follows FROM notification_settings WHERE user_id IN (%s)",
		join(users, ","),
	)

	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	targets := make([]Target, 0)
	for rows.Next() {
		target := Target{}
		err = rows.Scan(&target.ID, &target.RoomFrequency, &target.Follows)
		if err != nil {
			continue
		}

		targets = append(targets, target)
	}

	return targets, nil
}

func (s *Settings) UpdateSettingsFor(user int, frequency Frequency, follows bool) error {
	stmt, err := s.db.Prepare("UPDATE notification_settings SET room_frequency = $1, follows = $2 WHERE user_id = $3;")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(frequency, follows, user)
	return err
}

func join(elems []int64, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return strconv.FormatInt(elems[0], 10)
	}

	res := strconv.FormatInt(elems[0], 10)
	for _, s := range elems[1:] {
		res += sep
		res += strconv.FormatInt(s, 10)
	}

	return res
}
