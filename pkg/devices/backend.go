package devices

import (
	"database/sql"
)

type DevicesBackend struct {
	db *sql.DB
}

func NewDevicesBackend(db *sql.DB) *DevicesBackend {
	return &DevicesBackend{
		db: db,
	}
}

func (db *DevicesBackend) AddDeviceForUser(id int, token string) error {
	stmt, err := db.db.Prepare("INSERT INTO devices (token, user_id) VALUES ($1, $2);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(token, id)
	if err != nil {
		return err
	}

	return nil
}

func (db *DevicesBackend) GetDevicesForUser(id int) ([]string, error) {
	stmt, err := db.db.Prepare("SELECT token FROM devices WHERE user_id = $1;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0)

	for rows.Next() {
		var device string
		err := rows.Scan(&device)
		if err != nil {
			return nil, err
		}

		result = append(result, device)
	}

	return result, nil
}
