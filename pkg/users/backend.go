package users

import (
	"database/sql"
	"strings"
)

type User struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Email       string `json:"email"`
}

type UserBackend struct {
	db *sql.DB
}

func NewUserBackend(db *sql.DB) *UserBackend {
	return &UserBackend{
		db: db,
	}
}

func (ub *UserBackend) FindByEmail(email string) (*User, error) {
	stmt, err := ub.db.Prepare("SELECT id, display_name, username, email FROM users WHERE email = $1;")
	if err != nil {
		return nil, err
	}

	user := &User{}
	err = stmt.QueryRow(email).Scan(&user.ID, &user.DisplayName, &user.Username, &user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ub *UserBackend) CreateUser(email string, displayName string, username string) (int, error) {
	stmt, err := ub.db.Prepare("INSERT INTO users (display_name, username, email) VALUES ($1, $2, $3) RETURNING id;")
	if err != nil {
		return 0, err
	}

	var id int
	err = stmt.QueryRow(displayName, strings.ToLower(username), email).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}
