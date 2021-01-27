package rooms

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
)

type RoomState struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Members []RoomMember `json:"members"`
}

type RoomMember struct {
	ID          int    `json:"id"`
	DisplayName string `json:"display_name"`
	Image       string `json:"image"`
}

type Endpoint struct {
	repository *Repository
	server     *Server
}

func NewEndpoint(repository *Repository, server *Server) *Endpoint {
	return &Endpoint{
		repository: repository,
		server:     server,
	}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/v1/rooms", e.rooms).Methods("GET")
	r.HandleFunc("/v1/signal", e.server.Signal).Methods("GET")

	return r
}

func (e *Endpoint) rooms(w http.ResponseWriter, r *http.Request) {
	rooms := make([]*RoomState, 0)

	e.repository.Map(func(room *Room) {

		members := make([]RoomMember, 0)
		room.MapMembers(func(member *Member) {
			members = append(members, RoomMember{
				ID:          member.id,
				DisplayName: member.name,
				Image:       member.image,
			})
		})

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
