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
	stmt, err := s.db.Prepare("SELECT user_id, room_frequency, follows, welcome_rooms FROM notification_settings WHERE user_id = $1;")
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
	return s.getSettings(
		"SELECT notification_settings.user_id, notification_settings.room_frequency, notification_settings.follows, notification_settings.welcome_rooms FROM notification_settings INNER JOIN followers ON (notification_settings.user_id = followers.follower) WHERE followers.user_id = $1",
		user,
	)
}

func (s *Settings) GetSettingsForRecentlyActiveUsers() ([]Target, error) {
	return s.getSettings(
		`SELECT notification_settings.user_id, notification_settings.room_frequency, notification_settings.follows, notification_settings.welcome_rooms FROM notification_settings
		INNER JOIN (
			SELECT user_id
		    FROM (
		        SELECT user_id FROM current_rooms
		        UNION
		        SELECT user_id FROM user_active_times WHERE last_active > (NOW() - INTERVAL '15 MINUTE')
			) foo GROUP BY user_id) active
		ON notification_settings.user_id = active.user_id
		INNER JOIN user_room_time ON user_room_time.user_id = active.user_id  WHERE seconds >= 36000 AND visibility = 'public' AND active.user_id NOT IN (1, 75, 962);`,
	)
}

func (s *Settings) GetSettingsForUsers(users []int64) ([]Target, error) {
	query := fmt.Sprintf(
		"SELECT notification_settings.user_id, notification_settings.room_frequency, notification_settings.follows, notification_settings.welcome_rooms FROM notification_settings WHERE user_id IN (%s)",
		join(users, ","),
	)

	return s.getSettings(query)
}

func (s *Settings) UpdateSettingsFor(user int, frequency Frequency, follows, welcomeRooms bool) error {
	stmt, err := s.db.Prepare("UPDATE notification_settings SET room_frequency = $1, follows = $2, welcome_rooms = $3 WHERE user_id = $4;")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(frequency, follows, welcomeRooms, user)
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

func (s *Settings) getSettings(query string, args ...interface{}) ([]Target, error) {
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	targets := make([]Target, 0)
	for rows.Next() {
		target := Target{}
		err = rows.Scan(&target.ID, &target.RoomFrequency, &target.Follows, &target.WelcomeRooms)
		if err != nil {
			continue
		}

		targets = append(targets, target)
	}

	return targets, nil
}
