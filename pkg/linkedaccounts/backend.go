package linkedaccounts

import "database/sql"

type Backend struct {
	db *sql.DB
}

func NewLinkedAccountsBackend(db *sql.DB) *Backend {
	return &Backend{
		db: db,
	}
}

func (pb *Backend) LinkTwitterProfile(user, profile int, token, secret, username string) error {
	stmt, err := pb.db.Prepare("INSERT INTO profiles (user_id, provider, profile_id, token, secret, username) VALUES ($1, $2, $3, $4, $5, $6);")
	if err != nil {
		return err
	}

	_, err = stmt.Exec(user, "twitter", profile, token, secret, username)
	if err != nil {
		return err
	}

	return nil
}