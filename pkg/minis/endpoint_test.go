package minis_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/http/middlewares"
	"github.com/soapboxsocial/soapbox/pkg/minis"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
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

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	sm := sessions.NewSessionManager(rdb)
	mw := middlewares.NewAuthenticationMiddleware(sm)

	auth := "12345"
	_ = sm.NewSession(auth, 1, 0)

	endpoint := minis.NewEndpoint(minis.NewBackend(db), mw, nil)

	rr := httptest.NewRecorder()
	handler := endpoint.Router()

	mock.ExpectPrepare("SELECT id, name, slug, image, size, description FROM minis").
		ExpectQuery().
		WillReturnRows(mock.NewRows([]string{"id", "name", "slug", "image", "size", "description"}).AddRow(1, "name", "slug", "image", 0, ""))

	reader := strings.NewReader("")

	req, err := http.NewRequest("GET", "/", reader)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Authorization", auth)

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

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	sm := sessions.NewSessionManager(rdb)
	mw := middlewares.NewAuthenticationMiddleware(sm)

	auth := "12345"
	_ = sm.NewSession(auth, 1, 0)

	endpoint := minis.NewEndpoint(minis.NewBackend(db), mw, nil)

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

	req.Header.Set("Authorization", auth)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestEndpoint_SaveScores(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	scores := make(minis.Scores)
	scores[1] = 10

	key := "12345"
	room := "1235"
	mini := 10
	keys := make(minis.AuthKeys)
	keys[key] = mini

	endpoint := minis.NewEndpoint(minis.NewBackend(db), middlewares.NewAuthenticationMiddleware(nil), keys)

	rr := httptest.NewRecorder()
	handler := endpoint.Router()

	mock.ExpectBegin()

	mock.
		ExpectPrepare("^INSERT (.+)").
		ExpectExec().
		WithArgs(mini, room, 1, 10).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	body, err := json.Marshal(scores)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/scores?token="+key+"&room="+room, bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}
