package groups

import (
	"context"
	"database/sql"
)

type Group struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image,omitempty"`
	GroupType   string `json:"group_type"`
}

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{
		db: db,
	}
}

func (b *Backend) CreateGroup(creator int, name, description, image, groupType string) (int, error) {
	ctx := context.Background()
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO groups (name, description, image, group_type) VALUES ($1, $2, $3, (SELECT id FROM group_types WHERE name = $4));",
		name, description, image, groupType,
	)

	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO group_members (group_id, user_id, role) VALUES ((SELECT id FROM groups WHERE name = $1), $2, $3);",
		name, creator, "admin",
	)

	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	row := tx.QueryRow("SELECT id FROM groups WHERE name = $1", name)

	var id int
	err = row.Scan(&id)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return 0, nil
	}

	return id, nil
}

func (b *Backend) GetGroupsForUser(user, limit, offset int) ([]*Group, error) {
	stmt, err := b.db.Prepare("SELECT groups.id, groups.name, groups.description, groups.image, group_types.name AS group_type FROM groups INNER JOIN group_members ON (groups.id = group_members.group_id) INNER JOIN group_types ON (groups.group_type = group_types.id) WHERE group_members.user_id = $1 LIMIT $2 OFFSET $3;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(user, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]*Group, 0)

	for rows.Next() {
		group := &Group{}

		err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.Image, &group.GroupType)
		if err != nil {
			return nil, err // @todo
		}

		result = append(result, group)
	}

	return result, nil
}

func (b *Backend) IsAdminForGroup(user, group int) (bool, error) {
	stmt, err := b.db.Prepare("SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ? AND role = ?;")
	if err != nil {
		return false, err
	}

	row := stmt.QueryRow(group, user, "admin")
	var count int
	err = row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 1, nil
}

func (b *Backend) InviteUsers(from, group int, users []int) error {
	return nil
}