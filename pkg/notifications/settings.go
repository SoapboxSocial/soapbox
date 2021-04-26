package notifications

import "database/sql"

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
	stmt, err := s.db.Prepare("SELECT user_id, room_frequency, follows FROM notification_settings INNER JOIN followers ON (notification_settings.user_id = followers.follower) WHERE followers.user_id = $1")
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

func (s *Settings) UpdateSettingsFor(user int, frequency Frequency, follows bool) error {
	stmt, err := s.db.Prepare("UPDATE notification_settings SET room_frequency = $1, follows = $2 WHERE user_id = $3;")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(frequency, follows, user)
	return err
}
