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

	"github.com/soapboxsocial/soapbox/mocks"

	"github.com/soapboxsocial/soapbox/pkg/images"
	"github.com/soapboxsocial/soapbox/pkg/login"
	"github.com/soapboxsocial/soapbox/pkg/mail"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
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

	endpoint := login.NewEndpoint(
		users.NewBackend(db),
		login.NewStateManager(rdb),
		sessions.NewSessionManager(rdb),
		mail.NewMailService(&sendgrid.Client{}),
		images.NewImagesBackend("/foo"),
		pubsub.NewQueue(rdb),
		mocks.NewMockSignInWithApple(ctrl),
		mocks.NewMockRoomServiceClient(ctrl),
		login.Config{
			RegisterWithEmailEnabled: true,
		},
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

	endpoint := login.NewEndpoint(
		users.NewBackend(db),
		state,
		sessions.NewSessionManager(rdb),
		mail.NewMailService(&sendgrid.Client{}),
		images.NewImagesBackend("/foo"),
		pubsub.NewQueue(rdb),
		mocks.NewMockSignInWithApple(ctrl),
		mocks.NewMockRoomServiceClient(ctrl),
		login.Config{
			RegisterWithEmailEnabled: true,
		},
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

func TestLoginEndpoint_RegistrationCompleted(t *testing.T) {
	db, _, err := sqlmock.New()
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

	m := mocks.NewMockRoomServiceClient(ctrl)
	sm := sessions.NewSessionManager(rdb)

	endpoint := login.NewEndpoint(
		users.NewBackend(db),
		login.NewStateManager(rdb),
		sm,
		mail.NewMailService(&sendgrid.Client{}),
		images.NewImagesBackend("/foo"),
		pubsub.NewQueue(rdb),
		mocks.NewMockSignInWithApple(ctrl),
		m,
		login.Config{
			RegisterWithEmailEnabled: true,
		},
	)

	rr := httptest.NewRecorder()
	handler := endpoint.Router()

	session := "1234"
	userID := 1

	err = sm.NewSession(session, userID, 0)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/register/completed", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", session)

	m.
		EXPECT().
		RegisterWelcomeRoom(gomock.Any(), gomock.Eq(&pb.RegisterWelcomeRoomRequest{UserId: int64(userID)})).
		Return(&pb.RegisterWelcomeRoomResponse{Id: "foo"}, nil)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		print(rr.Body.String())
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
