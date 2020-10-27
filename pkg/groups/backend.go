package groups

import (
	"context"
	"database/sql"
)

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{
		db: db,
	}
}

func (b *Backend) CreateGroup(creator int, name, bio, image, groupType string) (int, error) {
	ctx := context.Background()
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO groups (name, bio, image, group_type) VALUES ($1, $2, $3, (SELECT id FROM group_types WHERE name = $4));",
		name, bio, image, groupType,
	)

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO group_members (group_id, user_id, role) VALUES ((SELECT id FROM groups WHERE name = $1), $2, $3);",
		name, creator, "admin",
	)

	if err != nil {
		tx.Rollback()
		return 0, err
	}

	row := tx.QueryRow("SELECT id FROM groups WHERE name = $1", name)

	var id int
	err = row.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
