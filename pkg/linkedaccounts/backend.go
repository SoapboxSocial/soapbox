package linkedaccounts

import "database/sql"

type LinkedAccount struct {
	ID        int
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
	stmt, err := pb.db.Prepare("SELECT profile_id, token, secret, username FROM linked_accounts WHERE user_id = $1 AND provider = $2")
	if err != nil {
		return nil, err
	}

	account := &LinkedAccount{ID: user, Provider: "twitter"}

	row := stmt.QueryRow(user, "twitter")

	err = row.Scan(&account.ProfileID, &account.Token, &account.Secret, &account.Username)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (pb *Backend) GetAllTwitterProfilesForUsersNotRecommendedToAndNotFollowedBy(user int) ([]LinkedAccount, error) {
	query := `
		SELECT user_id, profile_id, token, secret, username FROM linked_accounts 
		WHERE user_id NOT IN (SELECT user_id FROM followers WHERE follower = $1) 
   		AND user_id NOT IN (SELECT recommendation FROM follow_recommendations WHERE user_id = $1) AND user_id != $1`

	stmt, err := pb.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(user)
	if err != nil {
		return nil, err
	}

	result := make([]LinkedAccount, 0)
	for rows.Next() {
		account := LinkedAccount{Provider: "twitter"}

		err := rows.Scan(&account.ID, &account.ProfileID, &account.Token, &account.Secret, &account.Username)
		if err != nil {
			continue
		}

		result = append(result, account)
	}

	return result, nil
}
