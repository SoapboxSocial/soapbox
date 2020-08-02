package users

import "database/sql"

type User struct {
	ID          int64
	DisplayName string
	Username    string
	Email       string
}

type UserBackend struct {
	db *sql.DB
}

func NewUserBackend(db *sql.DB) *UserBackend {
	return &UserBackend{
		db: db,
	}
}

// @todo check for sql injections

func (ub *UserBackend) FindByEmail(email string) (*User, error) {
	row := ub.db.QueryRow("SELECT id, display_name, username, email FROM users WHERE email = ?", email)

	user := &User{}
	err := row.Scan(user.ID, user.DisplayName, user.Username, user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ub *UserBackend) CreateUser(email string, displayName string, username string) (int64, error) {
	res, err := ub.db.Exec("INSERT INTO users (display_name, username, email) VALUES ($1, $2, $3)", displayName, username, email)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}
