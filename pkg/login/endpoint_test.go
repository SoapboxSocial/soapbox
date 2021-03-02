package login_test

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/sendgrid/sendgrid-go"

	"github.com/alicebob/miniredis"

	"github.com/soapboxsocial/soapbox/pkg/images"
	"github.com/soapboxsocial/soapbox/pkg/login"
	mocks "github.com/soapboxsocial/soapbox/pkg/login/internal/mock"
	"github.com/soapboxsocial/soapbox/pkg/mail"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestLoginEndpoint_LoginWithTestAccount(t *testing.T) {
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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockSignInWithApple(ctrl)

	endpoint := login.NewEndpoint(
		users.NewUserBackend(db),
		login.NewStateManager(rdb),
		sessions.NewSessionManager(rdb),
		mail.NewMailService(&sendgrid.Client{}),
		images.NewImagesBackend("/foo"),
		pubsub.NewQueue(rdb),
		m,
		mocks.NewMockRoomServiceClient(ctrl),
	)

	rr := httptest.NewRecorder()
	handler := endpoint.Router()

	mock.ExpectPrepare("^SELECT (.+)").ExpectQuery().
		WithArgs(login.TestEmail).
		WillReturnRows(mock.NewRows([]string{"count"}).FromCSVString("0"))

	reader := strings.NewReader("email=" + login.TestEmail)

	req, err := http.NewRequest("POST", "/start", reader)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestLoginEndpoint_PinSubmission(t *testing.T) {
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

	state := login.NewStateManager(rdb)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockSignInWithApple(ctrl)

	endpoint := login.NewEndpoint(
		users.NewUserBackend(db),
		state,
		sessions.NewSessionManager(rdb),
		mail.NewMailService(&sendgrid.Client{}),
		images.NewImagesBackend("/foo"),
		pubsub.NewQueue(rdb),
		m,
		mocks.NewMockRoomServiceClient(ctrl),
	)

	rr := httptest.NewRecorder()
	handler := endpoint.Router()

	token := "1234"
	pin := "123456"
	email := "test@apple.com"

	err = state.SetPinState(token, email, pin)
	if err != nil {
		t.Fatal(err)
	}

	mock.ExpectPrepare("^SELECT (.+)").ExpectQuery().
		WithArgs(email).
		WillReturnRows(mock.NewRows([]string{"id", "display_name", "username", "image", "bio", "email"}).FromCSVString("1,dean,dean,123.png,my bio,test@apple.com"))

	form := url.Values{}
	form.Add("pin", pin)
	form.Add("token", token)

	req, err := http.NewRequest("POST", "/pin", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
