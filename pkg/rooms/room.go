package rooms

import (
	"encoding/json"
	"log"
	"sync"

	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"

	"github.com/soapboxsocial/soapbox/pkg/pb"
)

type PeerRole string

const (
	OWNER    PeerRole = "owner"
	SPEAKER  PeerRole = "speaker"
	AUDIENCE PeerRole = "audience"
)

type Member struct {
	ID          int      `json:"id"`
	DisplayName string   `json:"display_name"`
	Image       string   `json:"image"`
	Role        PeerRole `json:"role"`
	IsMuted     bool     `json:"is_muted"`
}

type peer struct {
	transport    *sfu.WebRTCTransport
	messageQueue chan *pb.RoomEvent
	member       *Member
}

// RoomLegacy represents the a Soapbox room, tracking its state and its peers.
type RoomLegacy struct {
	mux sync.RWMutex

	id    int
	sfu   *sfu.SFU
	peers map[int]*peer
}

// NewRoom returns a room
func NewRoom(id int, s *sfu.SFU) *RoomLegacy {
	return &RoomLegacy{
		id:    id,
		sfu:   s,
		peers: make(map[int]*peer),
	}
}

// PeerCount returns the number of connected peers.
func (r *RoomLegacy) PeerCount() int {
	r.mux.RLock()
	defer r.mux.RUnlock()
	return len(r.peers)
}

// join adds a user to the session using a webrtc offer.
func (r *RoomLegacy) Join(id int) (*sfu.WebRTCTransport, *webrtc.SessionDescription, error) {
	me := sfu.MediaEngine{}
	me.RegisterDefaultCodecs()

	peer, err := r.sfu.NewWebRTCTransport(string(r.id), me)
	if err != nil {
		return nil, nil, err
	}

	_, err = peer.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio)
	if err != nil {
		return nil, nil, err
	}

	offer, err := peer.CreateOffer()
	if err != nil {
		return nil, nil, err
	}

	err = peer.SetLocalDescription(offer)
	if err != nil {
		return nil, nil, err
	}

	// Notify user of trickle candidates
	peer.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		data, err := json.Marshal(c.ToJSON())
		if err != nil {
			log.Printf("json marshal candidate error: %v\n", err)
			return
		}

		r.mux.RLock()
		peer, ok := r.peers[id]
		r.mux.RUnlock()
		if !ok {
			return
		}

		peer.messageQueue <- &pb.RoomEvent{Type: pb.RoomEvent_CANDIDATE, From: 0, Data: data}
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

		r.mux.RLock()
		peer, ok := r.peers[id]
		r.mux.RUnlock()
		if !ok {
			return
		}

		peer.messageQueue <- &pb.RoomEvent{Type: pb.RoomEvent_OFFER, From: 0, Data: []byte(offer.SDP)}
	})

	return peer, &offer, nil
}

func (r *RoomLegacy) notify(event *pb.RoomEvent) {
	//data, err := proto.Marshal(event)
	//if err != nil {
	//	//
	//	return
	//}

	r.mux.RLock()
	defer r.mux.RUnlock()

	for id, peer := range r.peers {
		if int64(id) == event.From {
			continue
		}

		// @todo think about marshaling beforehand
		peer.messageQueue <- event
	}
}

func (r *RoomLegacy) closePeer(id int) {
	r.mux.RLock()
	peer, ok := r.peers[id]
	r.mux.RUnlock()
	if !ok {
		return
	}

	err := peer.transport.Close()
	if err != nil {
		log.Printf("peer.Close error: %v\n", err)
	}

	close(peer.messageQueue)

	r.mux.Lock()
	delete(r.peers, id)
	r.mux.Unlock()

	// @todo notify manager
}
