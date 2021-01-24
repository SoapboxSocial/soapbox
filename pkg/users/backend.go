package users

import (
	"database/sql"
	"strings"
)

// SearchUser is used for our search engine.
type SearchUser struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
	Image       string `json:"image"`
	Bio         string `json:"bio"`
	Followers   int    `json:"followers"`
}

type User struct {
	ID          int     `json:"id"`
	DisplayName string  `json:"display_name"`
	Username    string  `json:"username"`
	Image       string  `json:"image"`
	Bio         string  `json:"bio"`
	Email       *string `json:"email,omitempty"`
}

type LinkedAccount struct {
	ID       uint64 `json:"id"`
	Provider string `json:"provider"`
	Username string `json:"username"`
}

// Profile represents the User for public profile usage.
// This means certain fields like `email` are omitted,
// and others are added like `follower_counts` and relationships.
type Profile struct {
	ID             int             `json:"id"`
	DisplayName    string          `json:"display_name"`
	Username       string          `json:"username"`
	Bio            string          `json:"bio"`
	Followers      int             `json:"followers"`
	Following      int             `json:"following"`
	FollowedBy     *bool           `json:"followed_by,omitempty"`
	IsFollowing    *bool           `json:"is_following,omitempty"`
	IsBlocked      *bool           `json:"is_blocked,omitempty"`
	Image          string          `json:"image"`
	CurrentRoom    *int            `json:"current_room,omitempty"`
	LinkedAccounts []LinkedAccount `json:"linked_accounts"`
}

type NotificationUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Image    string `json:"image"`
}

type UserBackend struct {
	db *sql.DB
}

func NewUserBackend(db *sql.DB) *UserBackend {
	return &UserBackend{
		db: db,
	}
}

func (ub *UserBackend) GetIDForUsername(username string) (int, error) {
	stmt, err := ub.db.Prepare("SELECT id FROM users WHERE username = $1;")
	if err != nil {
		return 0, err
	}

	var id int
	err = stmt.QueryRow(username).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (ub *UserBackend) GetUserByUsername(username string) (*User, error) {
	stmt, err := ub.db.Prepare("SELECT id, display_name, image, bio FROM users WHERE username = $1;")
	if err != nil {
		return nil, err
	}

	user := &User{}
	err = stmt.QueryRow(username).Scan(&user.ID, &user.DisplayName, &user.Image, &user.Bio)
	if err != nil {
		return nil, err
	}

	user.Username = username

	return user, nil
}

func (ub *UserBackend) GetUserForSearchEngine(id int) (*SearchUser, error) {
	query := `SELECT 
       id, display_name, username, image, bio,
       (SELECT COUNT(*) FROM followers WHERE user_id = id) AS followers FROM users WHERE id = $1;`

	stmt, err := ub.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	profile := &SearchUser{}
	err = stmt.QueryRow(id).Scan(
		&profile.ID,
		&profile.DisplayName,
		&profile.Username,
		&profile.Image,
		&profile.Bio,
		&profile.Followers,
	)

	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (ub *UserBackend) GetMyProfile(id int) (*Profile, error) {
	query := `SELECT 
       id, display_name, username, image, bio,
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
		&profile.Bio,
		&profile.Followers,
		&profile.Following,
	)
	if err != nil {
		return nil, err
	}

	accounts, err := ub.LinkedAccounts(id)
	if err == nil {
		profile.LinkedAccounts = accounts
	}

	return profile, nil
}

func (ub *UserBackend) ProfileByID(id, from int) (*Profile, error) {
	query := `SELECT 
       id, display_name, username, image, bio,
       (SELECT COUNT(*) FROM followers WHERE user_id = id) AS followers,
       (SELECT COUNT(*) FROM followers WHERE follower = id) AS following,
       (SELECT COUNT(*) FROM followers WHERE follower = id AND user_id = $1) AS followed_by,
       (SELECT COUNT(*) FROM followers WHERE follower = $1 AND user_id = id) AS is_following,
       (SELECT COUNT(*) FROM blocks WHERE user_id = $1 AND blocked = id) AS is_following FROM users WHERE id = $2;`

	stmt, err := ub.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	profile := &Profile{}

	var followedBy, isFollowing, isBlocked int
	err = stmt.QueryRow(from, id).Scan(
		&profile.ID,
		&profile.DisplayName,
		&profile.Username,
		&profile.Image,
		&profile.Bio,
		&profile.Followers,
		&profile.Following,
		&followedBy,
		&isFollowing,
		&isBlocked,
	)

	if err != nil {
		return nil, err
	}

	following := isFollowing == 1
	followed := followedBy == 1
	blocked := isBlocked == 1
	profile.IsFollowing = &following
	profile.FollowedBy = &followed
	profile.IsBlocked = &blocked

	accounts, err := ub.LinkedAccounts(id)
	if err == nil {
		profile.LinkedAccounts = accounts
	}

	return profile, nil
}

func (ub *UserBackend) NotificationUserFor(id int) (*NotificationUser, error) {
	query := `SELECT id, username, image FROM users WHERE id = $1;`

	stmt, err := ub.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	profile := &NotificationUser{}

	err = stmt.QueryRow(id).Scan(
		&profile.ID,
		&profile.Username,
		&profile.Image,
	)

	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (ub *UserBackend) FindByAppleID(id string) (*User, error) {
	stmt, err := ub.db.Prepare("SELECT id, display_name, username, image, bio, email FROM users INNER JOIN apple_authentication ON users.id = apple_authentication.user_id WHERE apple_authentication.apple_user = $1;")
	if err != nil {
		return nil, err
	}

	user := &User{}
	err = stmt.QueryRow(id).Scan(&user.ID, &user.DisplayName, &user.Username, &user.Image, &user.Bio, &user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ub *UserBackend) FindByID(id int) (*User, error) {
	stmt, err := ub.db.Prepare("SELECT id, display_name, username, image, bio, email FROM users WHERE id = $1;")
	if err != nil {
		return nil, err
	}

	user := &User{}
	err = stmt.QueryRow(id).Scan(&user.ID, &user.DisplayName, &user.Username, &user.Image, &user.Bio, &user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ub *UserBackend) FindByEmail(email string) (*User, error) {
	stmt, err := ub.db.Prepare("SELECT id, display_name, username, image, bio, email FROM users WHERE email = $1;")
	if err != nil {
		return nil, err
	}

	user := &User{}
	err = stmt.QueryRow(email).Scan(&user.ID, &user.DisplayName, &user.Username, &user.Image, &user.Bio, &user.Email)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (ub *UserBackend) CreateUser(email, displayName, bio, image, username string) (int, error) {
	stmt, err := ub.db.Prepare("INSERT INTO users (display_name, username, email, bio, image) VALUES ($1, $2, $3, $4, $5) RETURNING id;")
	if err != nil {
		return 0, err
	}

	var id int
	err = stmt.QueryRow(displayName, strings.ToLower(username), email, bio, image).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (ub *UserBackend) UpdateUser(id int, displayName, bio, image string) error {
	query := "UPDATE users SET display_name = $1, bio = $2, image = $3 WHERE id = $4;"

	stmt, err := ub.db.Prepare(query)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(displayName, bio, image, id)
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

func (ub *UserBackend) LinkedAccounts(id int) ([]LinkedAccount, error) {
	stmt, err := ub.db.Prepare("SELECT profile_id, username, provider FROM linked_accounts WHERE user_id = $1;")
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(id)
	if err != nil {
		return nil, err
	}

	result := make([]LinkedAccount, 0)

	for rows.Next() {
		linked := LinkedAccount{}

		err := rows.Scan(&linked.ID, &linked.Username, &linked.Provider)
		if err != nil {
			return nil, err // @todo
		}

		result = append(result, linked)
	}

	return result, nil
}
