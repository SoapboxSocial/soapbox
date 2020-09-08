package users

import (
	"database/sql"
	"strings"
)

type User struct {
	ID          int     `json:"id"`
	DisplayName string  `json:"display_name"`
	Username    string  `json:"username"`
	Image       string  `json:"image"`
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
	CurrentRoom *int   `json:"current_room,omitempty"`
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

func (ub *UserBackend) ProfileByID(id, from int) (*Profile, error) {
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
	stmt, err := ub.db.Prepare("SELECT id, display_name, username, image, email FROM users WHERE id = $1;")
	if err != nil {
		return nil, err
	}

	user := &User{}
	err = stmt.QueryRow(id).Scan(&user.ID, &user.DisplayName, &user.Username, &user.Image, &user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ub *UserBackend) FindByEmail(email string) (*User, error) {
	stmt, err := ub.db.Prepare("SELECT id, display_name, username, image, email FROM users WHERE email = $1;")
	if err != nil {
		return nil, err
	}

	user := &User{}
	err = stmt.QueryRow(email).Scan(&user.ID, &user.DisplayName, &user.Username, &user.Image, &user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ub *UserBackend) CreateUser(email, displayName, image, username string) (int, error) {
	stmt, err := ub.db.Prepare("INSERT INTO users (display_name, username, email, image) VALUES ($1, $2, $3, $4) RETURNING id;")
	if err != nil {
		return 0, err
	}

	var id int
	err = stmt.QueryRow(displayName, strings.ToLower(username), email, image).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (ub *UserBackend) UpdateUser(id int, displayName, image string) error {
	query := "UPDATE users SET display_name = $1"
	params := []interface{}{displayName}

	count := "$2"
	if image != "" {
		query += ", image = $2"
		count = "$3"
		params = append(params, image)
	}

	query += " WHERE id = " + count + ";"
	params = append(params, id)

	stmt, err := ub.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(params...)
	return err
}

func (ub *UserBackend) GetProfileImage(id int) (string, error) {
	stmt, err := ub.db.Prepare("SELECT image FROM users WHERE id = $1;")
	if err != nil {
		return "", err
	}

	r := stmt.QueryRow(id)

	var name string
	err = r.Scan(&name)
	if err != nil {
		return "", err
	}

	return name, err
}
