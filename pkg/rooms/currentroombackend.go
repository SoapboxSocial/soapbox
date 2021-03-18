package rooms

import (
	"database/sql"
)

type CurrentRoomBackend struct {
	db *sql.DB
}

func NewCurrentRoomBackend(db *sql.DB) *CurrentRoomBackend {
	return &CurrentRoomBackend{
		db: db,
	}
}

func (b *CurrentRoomBackend) GetCurrentRoomForUser(id int) (string, error) {
	stmt, err := b.db.Prepare("SELECT room FROM current_room WHERE user_id = $1;")
	if err != nil {
		return "", err
	}

	row := stmt.QueryRow(id)

	var room string
	err = row.Scan(&room)
	if err != nil {
		return "", err
	}

	return room, nil
}

func (b *CurrentRoomBackend) SetCurrentRoomForUser(user int, room string) error {
	// @TODO IF THERE IS A CONFLICT.
	stmt, err := b.db.Prepare("INSERT INTO current_room (user_id, room) VALUES ($1, $2);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user, room)
	return err
}

func (b *CurrentRoomBackend) RemoveCurrentRoomForUser(user int) error {
	stmt, err := b.db.Prepare("DELETE FROM current_room WHERE user_id = $1")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user)
	return err
}
