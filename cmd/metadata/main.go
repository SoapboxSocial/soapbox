package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"google.golang.org/grpc"

	"github.com/gorilla/handlers"

	"github.com/soapboxsocial/soapbox/pkg/metadata"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

func main() {
	db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=voicely password=voicely dbname=voicely sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	usersBackend := users.NewUserBackend(db)


	endpoint := metadata.NewEndpoint(usersBackend)
	router := endpoint.Router()

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

	log.Print(http.ListenAndServe(":8081", handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}
