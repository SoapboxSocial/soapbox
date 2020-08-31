package roomsv2

import (
	"log"
	"sync"

	"github.com/golang/protobuf/proto"
	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"github.com/pkg/errors"

	"github.com/ephemeral-networks/soapbox/pkg/pb"
)

// Room represents the a Soapbox room, tracking its state and its peers.
type Room struct {
	sync.RWMutex

	id    int
	sfu   *sfu.SFU
	peers map[int]*sfu.WebRTCTransport
	messageQueue map[int] chan *pb.RoomEvent
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

	c := make(chan *pb.RoomEvent, 100)

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
		c <- &pb.RoomEvent{}
	})

	r.Lock()
	r.peers[id] = peer
	r.messageQueue[id] = c
	r.Unlock()

	peer.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		r.closePeer(id)
	})

	return &answer, nil
}

func (r *Room) handle(id int, d *webrtc.DataChannel) {
	d.OnOpen(func() {
		r.RLock()
		c := r.messageQueue[id]
		r.RUnlock()

		go func() {
			for {
				msg, ok := <-c
				if !ok {
					return
				}

				data, err := proto.Marshal(msg)
				if err != nil {
					log.Printf("proto.Marshal error: %v\n", err)
					continue
				}

				err = d.Send(data)
				if err != nil {
					// @todo
				}
			}
		}()
	})

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

	close(r.messageQueue[id])

	delete(r.peers, id)
	delete(r.messageQueue, id)

	// @todo notify manager
}
