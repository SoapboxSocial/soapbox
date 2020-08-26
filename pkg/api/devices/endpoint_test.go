package devices_test

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

	"github.com/ephemeral-networks/soapbox/pkg/api/devices"
	auth "github.com/ephemeral-networks/soapbox/pkg/api/middleware"
	backend "github.com/ephemeral-networks/soapbox/pkg/devices"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestDevicesEndpoint_AddDevice(t *testing.T) {
	token := "123"
	session := 123
	reader := strings.NewReader("token=" + token)

	r, err := http.NewRequest("POST", "/v1/devices/add", reader)
	if err != nil {
		t.Fatal(err)
	}

	req := r.WithContext(auth.WithUserID(r.Context(), session))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	endpoint := devices.NewDevicesEndpoint(backend.NewDevicesBackend(db))

	mock.ExpectPrepare("^INSERT (.+)").ExpectExec().
		WithArgs(token, session).
		WillReturnResult(sqlmock.NewResult(1, 1))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(endpoint.AddDevice)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestDevicesEndpoint_AddDeviceFailsWithoutToken(t *testing.T) {
	req, err := http.NewRequest("POST", "/v1/devices/add", strings.NewReader("foo=bar"))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	endpoint := devices.NewDevicesEndpoint(backend.NewDevicesBackend(db))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(endpoint.AddDevice)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestDevicesEndpoint_AddDeviceWithBackendError(t *testing.T) {
	token := "123"
	session := 123
	reader := strings.NewReader("token=" + token)

	r, err := http.NewRequest("POST", "/v1/devices/add", reader)
	if err != nil {
		t.Fatal(err)
	}

	req := r.WithContext(auth.WithUserID(r.Context(), session))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	endpoint := devices.NewDevicesEndpoint(backend.NewDevicesBackend(db))

	mock.ExpectPrepare("^INSERT (.+)").ExpectExec().
		WithArgs(token, session).
		WillReturnError(errors.New("boom"))

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(endpoint.AddDevice)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}
}

func TestDevicesEndpoint_AddDeviceWithoutForm(t *testing.T) {
	req, err := http.NewRequest("POST", "/v1/devices/add", nil)
	if err != nil {
		t.Fatal(err)
	}

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	endpoint := devices.NewDevicesEndpoint(backend.NewDevicesBackend(db))
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(endpoint.AddDevice)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}
