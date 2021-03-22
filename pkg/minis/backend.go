package minis

import "database/sql"

type Backend struct {
	db *sql.DB
}

func NewBackend(db *sql.DB) *Backend {
	return &Backend{db: db}
}

func (b *Backend) ListMinis() ([]Mini, error) {
	query := `SELECT id, name, slug, image, size, description FROM minis`

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

		err := rows.Scan(&mini.ID, &mini.Name, &mini.Slug, &mini.Image, &mini.Size, &mini.Description)
		if err != nil {
			continue
		}

		result = append(result, mini)
	}

	return result, nil
}

func (b *Backend) GetMiniWithSlug(slug string) (*Mini, error) {
	stmt, err := b.db.Prepare("SELECT id, name, image, size, description FROM minis WHERE slug = $1;")
	if err != nil {
		return nil, err
	}

	mini := &Mini{}
	err = stmt.QueryRow(slug).Scan(&mini.ID, &mini.Name, &mini.Image, &mini.Size, mini.Description)
	if err != nil {
		return nil, err
	}

	mini.Slug = slug

	return nil, nil
}

func (b *Backend) GetMiniWithID(id int) (*Mini, error) {
	stmt, err := b.db.Prepare("SELECT name, image, slug, size, description FROM minis WHERE id = $1;")
	if err != nil {
		return nil, err
	}

	mini := &Mini{}
	err = stmt.QueryRow(id).Scan(&mini.Name, &mini.Image, &mini.Slug, &mini.Size, mini.Description)
	if err != nil {
		return nil, err
	}

	mini.ID = id

	return nil, nil
}
