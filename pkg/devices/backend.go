package devices

import (
	"database/sql"
)

type Device struct {
	ID     int
	Device string
}

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

func (db *DevicesBackend) GetDevicesForUser(id int) ([]Device, error) {
	stmt, err := db.db.Prepare("SELECT user_id, token FROM devices WHERE user_id = $1;")
	if err != nil {
		return nil, err
	}

	return db.executeFetchDevicesQuery(stmt, id)
}

func (db *DevicesBackend) FetchAllFollowerDevices(id int) ([]Device, error) {
	stmt, err := db.db.Prepare("SELECT devices.user_id as id, devices.token FROM devices INNER JOIN followers ON (devices.user_id = followers.follower) WHERE followers.user_id = $1;")
	if err != nil {
		return nil, err
	}

	return db.executeFetchDevicesQuery(stmt, id)
}

func (db *DevicesBackend) executeFetchDevicesQuery(stmt *sql.Stmt, args ...interface{}) ([]Device, error) {
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	result := make([]Device, 0)

	for rows.Next() {
		device := Device{}

		err := rows.Scan(&device.ID, &device.Device)
		if err != nil {
			return nil, err
		}

		result = append(result, device)
	}

	return result, nil
}
