package account

import "database/sql"

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{
		db: db,
	}
}

func (b *Backend) DeleteAccount(id int) error {
	stmt, err := b.db.Prepare("DELETE FROM users WHERE id = $1")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(id)
	return err
}
