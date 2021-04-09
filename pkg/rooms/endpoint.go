package rooms

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type RoomState struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Visibility string       `json:"visibility"`
	Members    []RoomMember `json:"members"`
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
	r.HandleFunc("/v1/rooms/{id}", e.room).Methods("GET")
	r.HandleFunc("/v1/signal", e.server.Signal).Methods("GET")

	return r
}

func (e *Endpoint) rooms(w http.ResponseWriter, r *http.Request) {
	rooms := make([]RoomState, 0)

	userID, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	e.repository.Map(func(room *Room) {
		if room.ConnectionState() == closed {
			return
		}

		if !e.server.canJoin(userID, room) {
			return
		}

		rooms = append(rooms, roomToRoomState(room))
	})

	err := httputil.JsonEncode(w, rooms)
	if err != nil {
		log.Printf("rooms error: %v\n", err)
	}
}

func (e *Endpoint) room(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	userID, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	room, err := e.repository.Get(params["id"])
	if err != nil {
		httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeNotFound, "not found")
		return
	}

	if !e.server.canJoin(userID, room) {
		httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeNotFound, "not found")
		return
	}

	err = httputil.JsonEncode(w, roomToRoomState(room))
	if err != nil {
		log.Printf("room error: %v\n", err)
	}
}

// roomToRoomState turns a room into a RoomState object.
func roomToRoomState(room *Room) RoomState {
	members := make([]RoomMember, 0)
	room.MapMembers(func(member *Member) {
		members = append(members, RoomMember{
			ID:          member.id,
			DisplayName: member.name,
			Image:       member.image,
		})
	})

	visibility := "public"
	if room.visibility == pb.Visibility_VISIBILITY_PRIVATE {
		visibility = "private"
	}

	return RoomState{
		ID:         room.id,
		Name:       room.name,
		Visibility: visibility,
		Members:    members,
	}
}
