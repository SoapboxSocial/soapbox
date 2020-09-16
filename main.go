package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sendgrid/sendgrid-go"

	devicesapi "github.com/soapboxsocial/soapbox/pkg/api/devices"
	"github.com/soapboxsocial/soapbox/pkg/api/login"
	"github.com/soapboxsocial/soapbox/pkg/api/middleware"
	usersapi "github.com/soapboxsocial/soapbox/pkg/api/users"
	"github.com/soapboxsocial/soapbox/pkg/devices"
	"github.com/soapboxsocial/soapbox/pkg/followers"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/images"
	"github.com/soapboxsocial/soapbox/pkg/indexer"
	"github.com/soapboxsocial/soapbox/pkg/mail"
	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
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

	queue := notifications.NewNotificationQueue(rdb)
	index := indexer.NewIndexerQueue(rdb)

	s := sessions.NewSessionManager(rdb)
	ub := users.NewUserBackend(db)
	fb := followers.NewFollowersBackend(db)

	client, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	search := users.NewSearchBackend(client)

	devicesBackend := devices.NewDevicesBackend(db)

	amw := middleware.NewAuthenticationMiddleware(s)

	r := mux.NewRouter()

	r.MethodNotAllowedHandler = http.HandlerFunc(httputil.NotAllowedHandler)
	r.NotFoundHandler = http.HandlerFunc(httputil.NotFoundHandler)

	loginRoutes := r.PathPrefix("/v1/login").Methods("POST").Subrouter()

	ib := images.NewImagesBackend("/cdn/images")
	ms := mail.NewMailService(sendgrid.NewSendClient(sendgrid_api))
	loginHandlers := login.NewLoginEndpoint(ub, s, ms, ib, index, rdb)
	loginRoutes.HandleFunc("/start", loginHandlers.Start)
	loginRoutes.HandleFunc("/pin", loginHandlers.SubmitPin)
	loginRoutes.HandleFunc("/register", loginHandlers.Register)

	userRoutes := r.PathPrefix("/v1/users").Subrouter()

	usersEndpoints := usersapi.NewUsersEndpoint(ub, fb, s, queue, ib, search, index, rooms.NewCurrentRoomBackend(rdb))
	userRoutes.HandleFunc("/{id:[0-9]+}", usersEndpoints.GetUserByID).Methods("GET")
	userRoutes.HandleFunc("/{id:[0-9]+}/followers", usersEndpoints.GetFollowersForUser).Methods("GET")
	userRoutes.HandleFunc("/{id:[0-9]+}/following", usersEndpoints.GetFollowedByForUser).Methods("GET")
	userRoutes.HandleFunc("/follow", usersEndpoints.FollowUser).Methods("POST")
	userRoutes.HandleFunc("/unfollow", usersEndpoints.UnfollowUser).Methods("POST")
	userRoutes.HandleFunc("/edit", usersEndpoints.EditUser).Methods("POST")
	userRoutes.HandleFunc("/search", usersEndpoints.Search).Methods("GET")

	userRoutes.Use(amw.Middleware)

	devicesRoutes := r.PathPrefix("/v1/devices").Subrouter()

	devicesEndpoint := devicesapi.NewDevicesEndpoint(devicesBackend)
	devicesRoutes.HandleFunc("/add", devicesEndpoint.AddDevice).Methods("POST")
	devicesRoutes.Use(amw.Middleware)

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
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk)(r)))
}
