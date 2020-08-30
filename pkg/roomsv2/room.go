package roomsv2

import (
	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/ion-sfu/pkg/log"
	"github.com/pion/webrtc/v3"
	"github.com/pkg/errors"
)

// Room represents the a Soapbox room, tracking its state and its peers.
type Room struct {
	id    int
	sfu   *sfu.SFU
	peers map[int]string
}

// NewRoom returns a room
func NewRoom(id int, sfu *sfu.SFU) *Room {
	return &Room{
		id:    id,
		sfu:   sfu,
		peers: make(map[int]string),
	}
}

// Join adds a user to the session using a webrtc offer.
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
		log.Debugf("on negotiation needed called")
		offer, err := peer.CreateOffer()
		if err != nil {
			log.Errorf("CreateOffer error: %v", err)
			return
		}

		err = peer.SetLocalDescription(offer)
		if err != nil {
			log.Errorf("SetLocalDescription error: %v", err)
			return
		}

		// @todo
	})

	peer.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		// @todo
	})

	r.peers[id] = peer.ID()

	// @todo probably need to do onConnectionState change stuff to remove peers.

	return &answer, nil
}
