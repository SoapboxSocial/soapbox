package account_test

import (
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
	"github.com/golang/mock/gomock"

	"github.com/soapboxsocial/soapbox/pkg/account"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

func TestMain(m *testing.M) {
	log.SetOutput(ioutil.Discard)
	os.Exit(m.Run())
}

func TestAccountEndpoint_Delete(t *testing.T) {
	db, smock, err := sqlmock.New()
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

	sm := sessions.NewSessionManager(rdb)

	endpoint := account.NewEndpoint(
		account.NewBackend(db),
		pubsub.NewQueue(rdb),
		sm,
	)

	rr := httptest.NewRecorder()
	handler := endpoint.Router()

	session := "1234"
	userID := 1

	err = sm.NewSession(session, users.User{ID: userID}, 0)
	if err != nil {
		t.Fatal(err)
	}

	r, err := http.NewRequest("DELETE", "/", strings.NewReader(""))
	if err != nil {
		t.Fatal(err)
	}

	req := r.WithContext(httputil.WithUserID(r.Context(), userID))
	req.Header.Set("Authorization", session)

	smock.ExpectPrepare("^DELETE (.+)").
		ExpectExec().
		WithArgs(userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		print(rr.Body.String())
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
