package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
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

	log.Print(http.ListenAndServe(":8081", httputil.CORS(router)))
}
