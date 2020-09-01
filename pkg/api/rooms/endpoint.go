package rooms

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	sfu "github.com/pion/ion-sfu/pkg"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/roomsv2"
)

type Member struct {

}

type RoomPayload struct {
	ID      int            `json:"id"`
	Name    string         `json:"name,omitempty"`
	Members []Member `json:"members"`
}

type RoomsEndpoint struct {
	room     *roomsv2.Room
	upgrader *websocket.Upgrader
}

func NewRoomsEndpoint(sfu *sfu.SFU) RoomsEndpoint {
	return RoomsEndpoint{
		room:     roomsv2.NewRoom(1, sfu),
		upgrader: &websocket.Upgrader{},
	}
}

func (r *RoomsEndpoint) List(w http.ResponseWriter, req *http.Request) {
	//data := make([]RoomPayload, 0)
	//
	//manager.MapRooms(func(room *rooms.Room) {
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
		log.Print(err)
		// @todo
		return
	}

	//id, ok := middleware.GetUserIDFromContext(req.Context())
	//if !ok {
	//	// @todo error
	//	return
	//}
	//

	id := r.room.PeerCount()

	r.room.Handle(id, conn)
}
