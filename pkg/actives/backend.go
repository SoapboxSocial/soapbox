package actives

import "database/sql"

type ActiveUser struct {
	DisplayName string
	Username    string
	Image       string
	Room        string
	Stories     int
}

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{db: db}
}

func (b *Backend) GetActivesFor(user int) ([]ActiveUser, error) {
	return nil, nil
}
