package roomsv2

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"github.com/pkg/errors"
)

type Peer struct {
	transport  *sfu.WebRTCTransport
	connection *websocket.Conn
}

// Room represents the a Soapbox room, tracking its state and its peers.
type Room struct {
	sync.RWMutex

	id    int
	sfu   *sfu.SFU
	peers map[int]*Peer
}

// NewRoom returns a room
func NewRoom(id int, sfu *sfu.SFU) *Room {
	return &Room{
		id:    id,
		sfu:   sfu,
		peers: make(map[int]*Peer),
	}
}

// Handle a peers connection
func (r *Room) Handle(id int, c *websocket.Conn) {
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("ReadMessage error: %v\n", err)
			r.closePeer(id)
			return
		}

		// @todo decode message
		log.Println(message)
	}
}

// onJoin adds a user to the session using a webrtc offer.
// @TODO: probably pass message, and instead of conn put it into an array?
func (r *Room) onJoin(id int, c *websocket.Conn, offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	peer, err := r.sfu.NewWebRTCTransport(string(r.id), offer)
	if err != nil {
		return nil, errors.Wrap(err, "join error")
	}

	err = peer.SetRemoteDescription(offer)
	if err != nil {
		return nil, errors.Wrap(err, "join error")
	}

	answer, err := peer.CreateAnswer()
	if err != nil {
		return nil, errors.Wrap(err, "join error")
	}

	err = peer.SetLocalDescription(answer)
	if err != nil {
		return nil, errors.Wrap(err, "join error")
	}

	// Notify user of trickle candidates
	peer.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		// @todo
	})

	peer.OnNegotiationNeeded(func() {
		log.Println("on negotiation needed called")
		offer, err := peer.CreateOffer()
		if err != nil {
			log.Printf("CreateOffer error: %v\n", err)
			return
		}

		err = peer.SetLocalDescription(offer)
		if err != nil {
			log.Printf("SetLocalDescription error: %v\n", err)
			return
		}

		// @todo
	})

	// @TODO
	//peer.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
	//	r.Lock()
	//	delete(r.peers, id)
	//	r.Unlock()
	//
	//	err := peer.Close()
	//	if err != nil {
	//		log.Printf("peer.Close error: %v\n", err)
	//	}
	//})

	r.peers[id] = &Peer{
		connection: c,
		transport: peer,
	}

	return &answer, nil
}

func (r *Room) onAnswer(id int, desc webrtc.SessionDescription) {
	// @todo handle error
	_ = r.peers[id].transport.SetRemoteDescription(desc)
}

func (r *Room) closePeer(id int) {

}
