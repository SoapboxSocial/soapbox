package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/http/middlewares"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
)

func TestAuthenticationHandler_WithoutAuth(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	sm := sessions.NewSessionManager(rdb)
	mw := middlewares.NewAuthenticationMiddleware(sm)

	r, err := http.NewRequest("POST", "/add", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := mw.Middleware(nil)

	handler.ServeHTTP(rr, r)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestAuthenticationHandler_WithoutActiveSession(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	sm := sessions.NewSessionManager(rdb)
	mw := middlewares.NewAuthenticationMiddleware(sm)

	r, err := http.NewRequest("POST", "/add", nil)
	if err != nil {
		t.Fatal(err)
	}

	r.Header.Set("Authorization", "123")

	rr := httptest.NewRecorder()
	handler := mw.Middleware(nil)

	handler.ServeHTTP(rr, r)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestAuthenticationHandler(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	sm := sessions.NewSessionManager(rdb)
	mw := middlewares.NewAuthenticationMiddleware(sm)

	r, err := http.NewRequest("POST", "/add", nil)
	if err != nil {
		t.Fatal(err)
	}

	sess := "123"
	sm.NewSession(sess, 1, 0)

	r.Header.Set("Authorization", sess)

	rr := httptest.NewRecorder()
	handler := mw.Middleware(http.NotFoundHandler())

	handler.ServeHTTP(rr, r)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}
