package worker_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alicebob/miniredis"
	"github.com/dghubble/oauth1"
	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/linkedaccounts"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows/providers"
	"github.com/soapboxsocial/soapbox/pkg/recommendations/follows/worker"
)

// testServer returns an http Client, ServeMux, and Server. The client proxies
// requests to the server and handlers can be registered on the mux to handle
// requests. The caller must close the test server.
func testServer() (*http.Client, *http.ServeMux, *httptest.Server) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	transport := &RewriteTransport{&http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}}
	client := &http.Client{Transport: transport}
	return client, mux, server
}

// RewriteTransport rewrites https requests to http to avoid TLS cert issues
// during testing.
type RewriteTransport struct {
	Transport http.RoundTripper
}

// RoundTrip rewrites the request scheme to http and calls through to the
// composed RoundTripper or if it is nil, to the http.DefaultTransport.
func (t *RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	if t.Transport == nil {
		return http.DefaultTransport.RoundTrip(req)
	}
	return t.Transport.RoundTrip(req)
}

func TestWorker(t *testing.T) {
	httpClient, mux, server := testServer()
	defer server.Close()

	mux.HandleFunc("/1.1/friendships/lookup.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `[{"name": "andy piper (pipes)","screen_name": "andypiper","id": 1234,"id_str": "1234","connections": ["following"]}]`)
	})

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	pool := make(chan chan *worker.Job)

	w := worker.NewWorker(
		pool,
		&worker.Config{
			Twitter: providers.NewTwitter(
				&oauth1.Config{},
				linkedaccounts.NewLinkedAccountsBackend(db),
				httpClient,
			),
			Recommendations: follows.NewBackend(db),
			Queue:           pubsub.NewQueue(rdb),
		},
	)

	user := 1234

	timestamp := time.Now().Add(-(15 * (24 * time.Hour)))

	mock.
		ExpectPrepare("SELECT last_recommended (.+)").
		ExpectQuery().
		WithArgs(user).
		WillReturnRows(sqlmock.NewRows([]string{"last_follow_recommended"}).AddRow(timestamp))

	mock.
		ExpectPrepare("^SELECT (.+)").
		ExpectQuery().
		WithArgs(user, "twitter").
		WillReturnRows(sqlmock.NewRows([]string{"profile_id", "token", "secret", "username"}).AddRow(user, "foo", "bar", "baz"))

	respondedUser := 1

	mock.
		ExpectPrepare("^SELECT (.+)").
		ExpectQuery().
		WithArgs(user).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "profile_id", "token", "secret", "username"}).AddRow(respondedUser, 1234, "foo", "bar", "baz"))

	mock.
		ExpectBegin()

	mock.ExpectPrepare("^INSERT (.+)").
		ExpectExec().
		WithArgs(user, respondedUser).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	w.Start()

	queue := <-pool

	wg := &sync.WaitGroup{}

	wg.Add(1)
	queue <- &worker.Job{UserID: user, WaitGroup: wg}
	wg.Wait()
}
