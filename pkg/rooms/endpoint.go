package rooms

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
)

type RoomState struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Endpoint struct {
	repository *Repository
}

func NewEndpoint(repository *Repository) *Endpoint {
	return &Endpoint{repository: repository}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/rooms", e.rooms).Methods("GET")
	//r.HandleFunc("/rooms/{id:[0-9]+}", e.room).Methods("GET")

	return r
}

func (e *Endpoint) rooms(w http.ResponseWriter, r *http.Request) {
	rooms := make([]*RoomState, 0)

	// @TODO ACCESS TOKEN AND ALL THAT

	e.repository.Map(func(room *Room) {
		rooms = append(rooms, &RoomState{
			ID:   room.id,
			Name: room.name,
		})
	})

	err := httputil.JsonEncode(w, rooms)
	if err != nil {
		log.Printf("rooms error: %v\n", err)
	}
}
