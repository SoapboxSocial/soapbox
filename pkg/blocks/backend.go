package blocks

import (
	"database/sql"
)

type Backend struct {
	db *sql.DB
}

func (b *Backend) BlockUser(user, block int) error {
	stmt, err := b.db.Prepare("INSERT INTO blocks (user_id, blocked) VALUES ($1, $2);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user, block)
	if err != nil {
		return err
	}

	return nil
}

func (b *Backend) UnblockUser(user, block int) error {
	stmt, err := b.db.Prepare("DELETE FROM blocks WHERE user_id = $1 AND blocked = $2;")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user, block)
	if err != nil {
		return err
	}

	return nil
}

func (b *Backend) GetUsersWhoBlocked(user int) error {
	return nil
}

func (b *Backend) GetUsersBlockedBy(user int) error {
	return nil
}
