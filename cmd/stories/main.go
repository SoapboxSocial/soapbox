package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"

	"github.com/soapboxsocial/soapbox/pkg/stories"
)

func main() {

	db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=voicely password=voicely dbname=voicely sslmode=disable")
	if err != nil {
		panic(err)
	}

	backend := stories.NewBackend(db)
	files := stories.NewFileBackend("/cdn/stories")

	now := time.Now().Unix()

	ids, err := backend.DeleteExpired(now)
	if err != nil {
		panic(err)
	}

	for _, id := range ids {
		err := files.Remove(id + ".aac")
		if err != nil {
			log.Printf("files.Remove err: %v\n", err)
		}
	}
}
