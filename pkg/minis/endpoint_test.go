package minis_test

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/soapboxsocial/soapbox/pkg/minis"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestEndpoint_ListMinis(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	endpoint := minis.NewEndpoint(minis.NewBackend(db))

	rr := httptest.NewRecorder()
	handler := endpoint.Router()

	mock.ExpectPrepare("SELECT id, name, slug, image, description FROM minis").
		ExpectQuery().
		WillReturnRows(mock.NewRows([]string{"id", "name", "slug", "image", "description"}).AddRow(1, "name", "slug", "image", ""))

	reader := strings.NewReader("")

	req, err := http.NewRequest("GET", "/", reader)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestEndpoint_ListMinis_Error(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	endpoint := minis.NewEndpoint(minis.NewBackend(db))

	rr := httptest.NewRecorder()
	handler := endpoint.Router()

	mock.ExpectPrepare("SELECT id, name, slug, image, description FROM minis").
		ExpectQuery().
		WillReturnError(errors.New("rip"))

	reader := strings.NewReader("")

	req, err := http.NewRequest("GET", "/", reader)
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}
