package main

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/dghubble/oauth1"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sendgrid/sendgrid-go"

	"github.com/soapboxsocial/soapbox/pkg/activeusers"
	"github.com/soapboxsocial/soapbox/pkg/api/me"
	"github.com/soapboxsocial/soapbox/pkg/api/middleware"
	usersapi "github.com/soapboxsocial/soapbox/pkg/api/users"
	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/followers"
	"github.com/soapboxsocial/soapbox/pkg/groups"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/images"
	"github.com/soapboxsocial/soapbox/pkg/linkedaccounts"
	"github.com/soapboxsocial/soapbox/pkg/login"
	"github.com/soapboxsocial/soapbox/pkg/mail"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/search"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/stories"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

// @todo do this in config
const sendgrid_api = "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"

// @TODO: THINK ABOUT CHANGING QUEUES TO REDIS PUBSUB

func main() {
	db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=voicely password=voicely dbname=voicely sslmode=disable")
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	queue := pubsub.NewQueue(rdb)

	s := sessions.NewSessionManager(rdb)
	ub := users.NewUserBackend(db)
	fb := followers.NewFollowersBackend(db)
	ns := notifications.NewStorage(rdb)
	activeUsersBackend := activeusers.NewBackend(rdb, db)

	client, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	devicesBackend := devices.NewBackend(db)

	amw := middleware.NewAuthenticationMiddleware(s)

	r := mux.NewRouter()

	r.MethodNotAllowedHandler = http.HandlerFunc(httputil.NotAllowedHandler)
	r.NotFoundHandler = http.HandlerFunc(httputil.NotFoundHandler)

	ib := images.NewImagesBackend("/cdn/images")
	ms := mail.NewMailService(sendgrid.NewSendClient(sendgrid_api))

	loginEndpoints := login.NewEndpoint(ub, s, ms, ib, queue)
	loginRouter := loginEndpoints.Router()
	mount(r, "/v1/login", loginRouter)

	userRoutes := r.PathPrefix("/v1/users").Subrouter()

	usersEndpoints := usersapi.NewUsersEndpoint(
		ub,
		fb,
		s,
		ib,
		queue,
		rooms.NewCurrentRoomBackend(rdb),
		activeUsersBackend,
	)

	groupsBackend := groups.NewBackend(db)
	groupsEndpoint := groups.NewEndpoint(groupsBackend, ib, queue)

	storiesEndpoint := stories.NewEndpoint(stories.NewBackend(db), stories.NewFileBackend("/cdn/stories"), queue)
	storiesRouter := storiesEndpoint.Router()
	storiesRouter.Use(amw.Middleware)
	mount(r, "/v1/stories", storiesRouter)

	userRoutes.HandleFunc("/{id:[0-9]+}", usersEndpoints.GetUserByID).Methods("GET")
	userRoutes.HandleFunc("/{id:[0-9]+}/followers", usersEndpoints.GetFollowersForUser).Methods("GET")
	userRoutes.HandleFunc("/{id:[0-9]+}/following", usersEndpoints.GetFollowedByForUser).Methods("GET")
	userRoutes.HandleFunc("/friends", usersEndpoints.GetMyFriends).Methods("GET")
	userRoutes.HandleFunc("/follow", usersEndpoints.FollowUser).Methods("POST")
	userRoutes.HandleFunc("/unfollow", usersEndpoints.UnfollowUser).Methods("POST")
	userRoutes.HandleFunc("/edit", usersEndpoints.EditUser).Methods("POST")
	userRoutes.HandleFunc("/active", usersEndpoints.GetActiveUsersFor).Methods("GET")
	userRoutes.HandleFunc("/{id:[0-9]+}/groups", groupsEndpoint.GetGroupsForUser).Methods("GET")
	userRoutes.HandleFunc("/{id:[0-9]+}/stories", storiesEndpoint.GetStoriesForUser).Methods("GET")

	userRoutes.Use(amw.Middleware)

	devicesEndpoint := devices.NewEndpoint(devicesBackend)
	devicesRoutes := devicesEndpoint.Router()
	devicesRoutes.Use(amw.Middleware)
	mount(r, "/v1/devices", devicesRoutes)

	meRoutes := r.PathPrefix("/v1/me").Subrouter()

	// twitter oauth config
	oauth := oauth1.NewConfig(
		"nAzgMi6loUf3cl0hIkkXhZSth",
		"sFQEQ2cjJZSJgepUMmNyeTxiGggFXA1EKfSYAXpbARTu3CXBQY",
	)

	pb := linkedaccounts.NewLinkedAccountsBackend(db)

	meEndpoint := me.NewMeEndpoint(ub, groupsBackend, ns, oauth, pb)
	meRoutes.HandleFunc("", meEndpoint.GetMe).Methods("GET")
	meRoutes.HandleFunc("/notifications", meEndpoint.GetNotifications).Methods("GET")
	meRoutes.HandleFunc("/profiles/twitter", meEndpoint.AddTwitter).Methods("POST")
	meRoutes.HandleFunc("/profiles/twitter", meEndpoint.RemoveTwitter).Methods("DELETE")
	meRoutes.Use(amw.Middleware)

	groupsRouter := groupsEndpoint.Router()
	groupsRouter.Use(amw.Middleware)
	mount(r, "/v1/groups", groupsRouter)

	searchEndpoint := search.NewEndpoint(client)
	searchRouter := searchEndpoint.Router()
	searchRouter.Use(amw.Middleware)
	mount(r, "/v1/search", searchRouter)

	headersOk := handlers.AllowedHeaders([]string{
		"Content-Type",
		"X-Requested-With",
		"Accept",
		"Accept-Language",
		"Accept-Encoding",
		"Content-Language",
		"Origin",
	})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk)(r)))
}

func mount(r *mux.Router, path string, handler http.Handler) {
	r.PathPrefix(path).Handler(
		http.StripPrefix(
			strings.TrimSuffix(path, "/"),
			AddSlashForRoot(handler),
		),
	)
}

// AddSlashForRoot adds a slash if the path is the root path.
// This is necessary for our subrouters where there may be a root.
func AddSlashForRoot(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// @TODO MAYBE ENSURE SUFFIX DOESN'T ALREADY EXIST?
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}

		next.ServeHTTP(w, r)
	})
}
