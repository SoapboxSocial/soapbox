package rooms

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	sfu "github.com/pion/ion-sfu/pkg"

	"github.com/soapboxsocial/soapbox/pkg/api/middleware"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Member struct {
}

type RoomPayload struct {
	ID      int      `json:"id"`
	Name    string   `json:"name,omitempty"`
	Members []Member `json:"members"`
}

type RoomsEndpoint struct {
	room     *rooms.RoomLegacy
	upgrader *websocket.Upgrader
	ub       *users.UserBackend
}

func NewRoomsEndpoint(sfu *sfu.SFU, ub *users.UserBackend) RoomsEndpoint {
	return RoomsEndpoint{
		room:     rooms.NewRoom(1, sfu),
		upgrader: &websocket.Upgrader{},
		ub:       ub,
	}
}

func (r *RoomsEndpoint) List(w http.ResponseWriter, req *http.Request) {
	//data := make([]RoomPayload, 0)
	//
	//manager.MapRooms(func(room *rooms.RoomLegacy) {
	//	if room == nil {
	//		return
	//	}
	//
	//	r := RoomPayload{ID: room.GetID(), Members: make([]rooms.Member, 0)}
	//
	//	name := room.GetName()
	//	if name != "" {
	//		r.Name = name
	//	}
	//
	//	room.MapPeers(func(id int, peer rooms.Peer) {
	//		r.Members = append(r.Members, peer.GetMember())
	//	})
	//
	//	data = append(data, r)
	//})

	data := []RoomPayload{
		{ID: 1, Name: "Pew", Members: make([]Member, 0)},
	}

	err := httputil.JsonEncode(w, data)
	if err != nil {
		fmt.Println(err)
	}
}

func (r *RoomsEndpoint) Join(w http.ResponseWriter, req *http.Request) {
	conn, err := r.upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println(err)
		// @todo
		return
	}

	id, ok := middleware.GetUserIDFromContext(req.Context())
	if !ok {
		// @todo error
		return
	}

	user, err := r.ub.FindByID(id)
	if err != nil {
		log.Println(err)
		conn.Close()
		return
	}

	r.room.Handle(user, conn)
}
