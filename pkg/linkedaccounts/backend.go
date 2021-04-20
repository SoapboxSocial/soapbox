package linkedaccounts

import "database/sql"

type LinkedAccount struct {
	Provider  string
	ProfileID int64
	Token     string
	Secret    string
	Username  string
}

type Backend struct {
	db *sql.DB
}

func NewLinkedAccountsBackend(db *sql.DB) *Backend {
	return &Backend{
		db: db,
	}
}

func (pb *Backend) LinkTwitterProfile(user, profile int, token, secret, username string) error {
	stmt, err := pb.db.Prepare("INSERT INTO linked_accounts (user_id, provider, profile_id, token, secret, username) VALUES ($1, $2, $3, $4, $5, $6);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user, "twitter", profile, token, secret, username)
	if err != nil {
		return err
	}

	return nil
}

func (pb *Backend) UnlinkTwitterProfile(user int) error {
	stmt, err := pb.db.Prepare("DELETE FROM linked_accounts WHERE user_id = $1 AND provider = $2;")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user, "twitter")
	if err != nil {
		return err
	}

	return nil
}

func (pb *Backend) GetTwitterProfileFor(user int) (*LinkedAccount, error) {
	return nil, nil
}

func (pb *Backend) GetAllTwitterProfilesForUsersNotFollowedBy(user int) ([]LinkedAccount, error) {
	return nil, nil
}
