package notifications

import "database/sql"

type Targets struct {
	db *sql.DB
}

func NewTargets(db *sql.DB) *Targets {
	return &Targets{db: db}
}

func (s *Targets) GetTargetFor(user int) (*Target, error) {
	stmt, err := s.db.Prepare("SELECT user_id, room_frequency, follows FROM notification_settings WHERE user_id = $1;")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(user)

	target := &Target{}
	err = row.Scan(&target.ID, &target.Frequency, &target.Follows)
	if err != nil {
		return nil, err
	}

	return target, nil
}

func (s *Targets) GetTargetsFollowingUser(user int) ([]Target, error) {
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
		err = rows.Scan(&target.ID, &target.Frequency, &target.Follows)
		if err != nil {
			continue
		}

		targets = append(targets, target)
	}

	return targets, nil
}
