package minis

import "database/sql"

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{db: db}
}

func (b *Backend) ListMinis() ([]Mini, error) {
	query := `SELECT id, name, slug, image, description FROM minis`

	stmt, err := b.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	result := make([]Mini, 0)

	for rows.Next() {
		mini := Mini{}

		err := rows.Scan(&mini.ID, &mini.Name, &mini.Slug, &mini.Image, &mini.Description)
		if err != nil {
			continue
		}

		result = append(result, mini)
	}

	return result, nil
}

func (b *Backend) GetMiniWithSlug(slug string) (*Mini, error) {
	return nil, nil
}


func (b *Backend) GetMiniWithID(id int64) (*Mini, error) {
	return nil, nil
}
