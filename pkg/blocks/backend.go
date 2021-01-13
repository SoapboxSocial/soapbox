package blocks

import (
	"database/sql"
)

type Backend struct {
	db *sql.DB
}

func (b *Backend) BlockUser(user, block int) error {
	return nil
}

func (b *Backend) UnblockUser(user, block int) error {
	return nil
}

func (b *Backend) GetUsersWhoBlocked(user int) error {
	return nil
}

func (b *Backend) GetUsersBlockedBy(user int) error {
	return nil
}
