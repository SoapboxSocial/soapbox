package users

import (
	"database/sql"
	"strings"
)

type User struct {
	ID          int     `json:"id"`
	DisplayName string  `json:"display_name"`
	Username    string  `json:"username"`
	Email       *string `json:"email,omitempty"`
}

// Profile represents the User for public profile usage.
// This means certain fields like `email` are omitted,
// and others are added like `follower_counts` and relationships.
type Profile struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Followers   int    `json:"followers"`
	Following   int    `json:"following"`
	FollowedBy  *bool  `json:"followed_by,omitempty"`
	IsFollowing *bool  `json:"is_following,omitempty"`
	Image       string `json:"image"`
}

type UserBackend struct {
	db *sql.DB
}

func NewUserBackend(db *sql.DB) *UserBackend {
	return &UserBackend{
		db: db,
	}
}

func (ub *UserBackend) GetMyProfile(id int) (*Profile, error) {
	query := `SELECT 
       id, display_name, username, image,
       (SELECT COUNT(*) FROM followers WHERE user_id = id) AS followers,
       (SELECT COUNT(*) FROM followers WHERE follower = id) AS following FROM users WHERE id = $1;`

	stmt, err := ub.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	profile := &Profile{}
	err = stmt.QueryRow(id).Scan(
		&profile.ID,
		&profile.DisplayName,
		&profile.Username,
		&profile.Image,
		&profile.Followers,
		&profile.Following,
	)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (ub *UserBackend) ProfileByID(id int, from int) (*Profile, error) {
	query := `SELECT 
       id, display_name, username, image,
       (SELECT COUNT(*) FROM followers WHERE user_id = id) AS followers,
       (SELECT COUNT(*) FROM followers WHERE follower = id) AS following,
       (SELECT COUNT(*) FROM followers WHERE follower = id AND user_id = $1) AS followed_by,
       (SELECT COUNT(*) FROM followers WHERE follower = $2 AND user_id = id) AS is_following FROM users WHERE id = $3;`

	stmt, err := ub.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	profile := &Profile{}

	var followedBy, isFollowing int
	err = stmt.QueryRow(from, from, id).Scan(
		&profile.ID,
		&profile.DisplayName,
		&profile.Username,
		&profile.Image,
		&profile.Followers,
		&profile.Following,
		&followedBy,
		&isFollowing,
	)

	if err != nil {
		return nil, err
	}

	following := isFollowing == 1
	followed := followedBy == 1
	profile.IsFollowing = &following
	profile.FollowedBy = &followed

	return profile, nil
}

func (ub *UserBackend) FindByID(id int) (*User, error) {
	stmt, err := ub.db.Prepare("SELECT id, display_name, username, email FROM users WHERE id = $1;")
	if err != nil {
		return nil, err
	}

	user := &User{}
	err = stmt.QueryRow(id).Scan(&user.ID, &user.DisplayName, &user.Username, &user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
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

func (ub *UserBackend) UpdateUser(id int, displayName string) error {
	stmt, err := ub.db.Prepare("UPDATE users SET display_name = $1 WHERE id = $2;")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(displayName, id)
	return err
}

func (ub *UserBackend) UpdateUserImage(id int, path string) error {
	stmt, err := ub.db.Prepare("UPDATE users SET image = $1 WHERE id = $2;")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(path, id)
	return err
}
