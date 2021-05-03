package analytics_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/soapboxsocial/soapbox/pkg/analytics"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestEndpoint_OpenedNotification(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	endpoint := analytics.NewEndpoint(
		analytics.NewBackend(db),
	)

	rr := httptest.NewRecorder()
	handler := endpoint.Router()

	session := "1234"
	userID := 1
	notification := "12345678"

	r, err := http.NewRequest("POST", fmt.Sprintf("/notifications/%s/opened", notification), strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}

	req := r.WithContext(httputil.WithUserID(r.Context(), userID))
	req.Header.Set("Authorization", session)

	mock.ExpectPrepare("^UPDATE (.+)").
		ExpectExec().
		WithArgs(userID, notification).
		WillReturnResult(sqlmock.NewResult(1, 1))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		print(rr.Body.String())
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
