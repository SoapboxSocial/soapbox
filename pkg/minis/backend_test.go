package minis_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/soapboxsocial/soapbox/pkg/minis"
)

func TestBackend_GetMiniWithID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	backend := minis.NewBackend(db)

	id := 1
	mock.ExpectPrepare("SELECT").
		ExpectQuery().
		WithArgs(id).
		WillReturnRows(mock.NewRows([]string{"name", "slug", "image", "size", "description"}).AddRow("name", "slug", "image", 0, ""))

	result, err := backend.GetMiniWithID(id)
	if err != nil {
		t.Fatal(err)
	}

	if result.ID != id {
		t.Fatal("id not matching")
	}
}

func TestBackend_GetMiniWithSlug(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	backend := minis.NewBackend(db)

	slug := "/1"
	mock.ExpectPrepare("SELECT").
		ExpectQuery().
		WithArgs(slug).
		WillReturnRows(mock.NewRows([]string{"name", "slug", "image", "size", "description"}).AddRow(1, "name", "image", 0, ""))

	result, err := backend.GetMiniWithSlug(slug)
	if err != nil {
		t.Fatal(err)
	}

	if result.Slug != slug {
		t.Fatal("slug not matching")
	}
}
