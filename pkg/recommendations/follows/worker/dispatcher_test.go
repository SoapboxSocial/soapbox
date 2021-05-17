package worker_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows/worker"
)

func TestDispatcher(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	dispatcher := worker.NewDispatcher(1, &worker.Config{Recommendations: follows.NewBackend(db)})

	dispatcher.Run()

	for i := 0; i < 30; i++ {
		expect := mock.ExpectPrepare("^SELECT (.+)").WillReturnError(errors.New("poop"))
		if i == 29 {
			expect.WillDelayFor(1 * time.Second)
		}

		dispatcher.Dispatch(i)
	}

	dispatcher.Wait()
	dispatcher.Stop()

	err = mock.ExpectationsWereMet()
	if err != nil {
		t.Fatal(err)
	}
}
