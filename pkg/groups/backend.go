package groups

import (
	"context"
	"database/sql"

	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Group struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Image       string `json:"image,omitempty"`
	GroupType   string `json:"group_type"`
	IsMember    *bool  `json:"is_member,omitempty"`
	IsInvited   *bool  `json:"is_invited,omitempty"`
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
		return 0, err
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

func (b *Backend) FindById(id int) (*Group, error) {
	stmt, err := b.db.Prepare("SELECT groups.id, groups.name, groups.description, groups.image, group_types.name AS group_type FROM groups INNER JOIN group_types ON (groups.group_type = group_types.id) WHERE groups.id = $1;")
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(id)

	group := &Group{}

	err = row.Scan(&group.ID, &group.Name, &group.Description, &group.Image, &group.GroupType)
	if err != nil {
		return nil, err // @todo
	}

	return group, nil
}

func (b *Backend) IsAdminForGroup(user, group int) (bool, error) {
	row := b.db.QueryRow(
		"SELECT COUNT(*) FROM group_members WHERE group_id = $1 AND user_id = $2 AND role = $3;",
		group, user, "admin",
	)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 1, nil
}

func (b *Backend) GetGroupForUser(user, groupId int) (*Group, error) {
	query := `SELECT
		groups.id, groups.name, groups.description, groups.image,
		(SELECT COUNT(*) FROM group_members WHERE group_id = $1 AND user_id = $2) AS is_member,
		(SELECT COUNT(*) FROM group_invites WHERE group_id = $1 AND user_id = $2) AS is_invited,
		group_types.name AS group_type FROM groups INNER JOIN group_types ON (groups.group_type = group_types.id) WHERE groups.id = $1;`

	stmt, err := b.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	group := &Group{}

	var isMember, isInvited int
	err = stmt.QueryRow(groupId, user).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.Image,
		&isMember,
		&isInvited,
		&group.GroupType,
	)

	if err != nil {
		return nil, err
	}

	var member = isMember == 1
	var invited = isInvited == 1
	group.IsMember = &member
	group.IsInvited = &invited

	return group, nil
}

func (b *Backend) GetInviterForUser(userId, groupId int) (*users.User, error) {
	stmt, err := b.db.Prepare("SELECT users.* FROM users INNER JOIN group_invites ON (users.id = group_invites.from_id) WHERE group_invites.user_id = $1 AND group_invites.group_id = $2;")
	if err != nil {
		return nil, err
	}

	user := &users.User{}
	err = stmt.QueryRow(userId, groupId).Scan(&user.ID, &user.DisplayName, &user.Username, &user.Image, &user.Bio, &user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (b *Backend) DeclineInvite(userId, groupId int) error {
	stmt, err := b.db.Prepare("DELETE FROM group_invites WHERE user_id = $1 AND group_id = $2")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(userId, groupId)
	return err
}

func (b *Backend) AcceptInvite(userId, groupId int) error {
	ctx := context.Background()
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	row := tx.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM group_invites WHERE user_id = $1 AND group_id = $2",
		userId, groupId,
	)

	var count int
	err = row.Scan(&count)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"DELETE FROM group_invites WHERE user_id = $1 AND group_id = $2",
		userId, groupId,
	)

	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = tx.ExecContext(
		ctx,
		"INSERT INTO group_members (group_id, user_id, role) VALUES ($1, $2, $3);",
		groupId, userId, "user",
	)

	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}

func (b *Backend) InviteUser(from, group, user int) error {
	stmt, err := b.db.Prepare("INSERT INTO group_invites (user_id, group_id, from_id) VALUES ($1, $2, $3);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user, group, from)
	if err != nil {
		return err
	}

	return nil
}

func (b *Backend) GetAllMembers(id int, limit, offset int) ([]*users.User, error) {
	stmt, err := b.db.Prepare("SELECT users.id, users.display_name, users.username, users.image FROM users INNER JOIN group_members ON (users.id = group_members.user_id) WHERE group_members.group = $1 LIMIT $2 OFFSET $3;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(id, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]*users.User, 0)

	for rows.Next() {
		user := &users.User{}

		err := rows.Scan(&user.ID, &user.DisplayName, &user.Username, &user.Image)
		if err != nil {
			return nil, err // @todo
		}

		result = append(result, user)
	}

	return result, nil
}
