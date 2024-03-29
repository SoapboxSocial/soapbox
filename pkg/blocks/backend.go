package blocks

import (
	"context"
	"database/sql"
)

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{db: db}
}

func (b *Backend) BlockUser(user, block int) error {
	ctx := context.Background()
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO blocks (user_id, blocked) VALUES ($1, $2);",
		user, block,
	)

	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"DELETE FROM followers WHERE (follower = $1 AND user_id = $2) OR (follower = $2 AND user_id = $1);",
		user, block,
	)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
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

func (b *Backend) GetUsersWhoBlocked(user int) ([]int, error) {
	stmt, err := b.db.Prepare("SELECT user_id FROM blocks WHERE blocked = $1;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(user)
	if err != nil {
		return nil, err
	}

	result := make([]int, 0)

	for rows.Next() {
		var blocker int
		err := rows.Scan(&blocker)
		if err != nil {
			return nil, err
		}

		result = append(result, blocker)
	}

	return result, nil
}

func (b *Backend) GetUsersBlockedBy(user int) ([]int, error) {
	stmt, err := b.db.Prepare("SELECT blocked FROM blocks WHERE user_id = $1;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(user)
	if err != nil {
		return nil, err
	}

	result := make([]int, 0)

	for rows.Next() {
		var blocker int
		err := rows.Scan(&blocker)
		if err != nil {
			return nil, err
		}

		result = append(result, blocker)
	}

	return result, nil
}
