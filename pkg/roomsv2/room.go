package roomsv2

import (
	"log"
	"sync"

	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"github.com/pkg/errors"
)

// Room represents the a Soapbox room, tracking its state and its peers.
type Room struct {
	sync.RWMutex

	id    int
	sfu   *sfu.SFU
	peers map[int]*sfu.WebRTCTransport
}

// NewRoom returns a room
func NewRoom(id int, s *sfu.SFU) *Room {
	return &Room{
		id:    id,
		sfu:   s,
		peers: make(map[int]*sfu.WebRTCTransport),
	}
}

// Join adds a user to the session using a webrtc offer.
// @TODO: probably pass message, and instead of conn put it into an array?
func (r *Room) Join(id int, offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
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

	r.Lock()
	r.peers[id] = peer
	r.Unlock()

	peer.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		r.closePeer(id)
	})

	return &answer, nil
}

func (r *Room) handle(id int, d *webrtc.DataChannel) {
	d.OnMessage(func(msg webrtc.DataChannelMessage) {

	})

	d.OnClose(func() {
		r.closePeer(id)
	})
}

func (r *Room) onAnswer(id int, desc webrtc.SessionDescription) {
	// @todo handle error
	_ = r.peers[id].SetRemoteDescription(desc)
}

func (r *Room) closePeer(id int) {
	r.Lock()
	defer r.Unlock()

	err := r.peers[id].Close()
	if err != nil {
		log.Printf("peer.Close error: %v\n", err)
	}

	delete(r.peers, id)

	// @todo notify manager
}
