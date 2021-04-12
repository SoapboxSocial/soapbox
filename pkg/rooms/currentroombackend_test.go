package rooms_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/soapboxsocial/soapbox/pkg/rooms"
)

func TestCurrentRoomBackend_GetCurrentRoomForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	backend := rooms.NewCurrentRoomBackend(db)

	user := 10
	room := "foo"

	mock.
		ExpectPrepare("SELECT room").
		ExpectQuery().
		WillReturnRows(sqlmock.NewRows([]string{"room"}).AddRow(room))

	val, err := backend.GetCurrentRoomForUser(user)
	if err != nil {
		t.Fatal(err)
	}

	if val != room {
		t.Fatalf("expected: %s actual: %s", room, val)
	}
}

func TestCurrentRoomBackend_SetCurrentRoomForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	backend := rooms.NewCurrentRoomBackend(db)

	user := 10
	room := "foo"

	mock.
		ExpectPrepare("SELECT").
		ExpectExec().
		WithArgs(user, room).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = backend.SetCurrentRoomForUser(user, room)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCurrentRoomBackend_RemoveCurrentRoomForUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	backend := rooms.NewCurrentRoomBackend(db)

	user := 10

	mock.
		ExpectPrepare("DELETE FROM current_rooms").
		ExpectExec().
		WithArgs(user).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = backend.RemoveCurrentRoomForUser(user)
	if err != nil {
		t.Fatal(err)
	}
}
