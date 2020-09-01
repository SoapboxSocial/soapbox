package roomsv2

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"github.com/pkg/errors"

	"github.com/soapboxsocial/soapbox/pkg/pb"
)

type peer struct {
	transport *sfu.WebRTCTransport
	messageQueue chan *pb.RoomEvent
}

// Room represents the a Soapbox room, tracking its state and its peers.
type Room struct {
	mux sync.RWMutex

	id    int
	sfu   *sfu.SFU
	peers map[int]*peer
}

// NewRoom returns a room
func NewRoom(id int, s *sfu.SFU) *Room {
	return &Room{
		id:    id,
		sfu:   s,
		peers: make(map[int]*peer),
	}
}

func (r *Room) PeerCount() int {
	r.mux.RLock()
	defer r.mux.RUnlock()
	return len(r.peers)
}

func (r *Room) Handle(id int, conn *websocket.Conn) {
	transport, offer, err := r.join(id, conn)
	if err != nil {
		log.Printf("failed to join: %v\n", err)
		_ = conn.Close()
	}

	r.mux.Lock()
	r.peers[id] = &peer{
		transport: transport,
		messageQueue: make(chan *pb.RoomEvent, 100),
	}
	r.mux.Unlock()

	event := &pb.RoomEvent{Type: pb.RoomEvent_OFFER, From: 0, Data: []byte(offer.SDP)}
	data, err := proto.Marshal(event)
	if err != nil {

	}

	err = conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		log.Printf("conn.WriteMessage error: %v\n", err)
	}

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			// @todo
			return
		}

		if mt != websocket.BinaryMessage {
			continue
		}

		cmd := &pb.RoomCommand{}
		err = proto.Unmarshal(message, cmd)
		if err != nil {
			// @todo
		}

		switch cmd.Type {
		case pb.RoomCommand_ANSWER:
			r.onAnswer(id, webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP: string(cmd.Data),
			})
		case pb.RoomCommand_CANDIDATE:
			candidate := &webrtc.ICECandidateInit{}
			err := json.Unmarshal(cmd.Data, candidate)
			if err != nil {
				// @todo
				continue
			}

			r.onCandidate(id, candidate)
		default:
			continue
		}
	}
}

// join adds a user to the session using a webrtc offer.
func (r *Room) join(id int, conn *websocket.Conn) (*sfu.WebRTCTransport, *webrtc.SessionDescription, error) {
	me := sfu.MediaEngine{}
	me.RegisterDefaultCodecs()

	peer, err := r.sfu.NewWebRTCTransport(string(r.id), me)
	if err != nil {
		return nil, nil, errors.Wrap(err, "join error")
	}

	_, err = peer.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio)
	if err != nil {
		return nil, nil, errors.Wrap(err, "join error")
	}

	offer, err := peer.CreateOffer()
	if err != nil {
		return nil, nil, errors.Wrap(err, "join error")
	}

	err = peer.SetLocalDescription(offer)
	if err != nil {
		return nil, nil, errors.Wrap(err, "join error")
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

		event := &pb.RoomEvent{Type: pb.RoomEvent_CANDIDATE, From: 0, Data: data}
		protobuf, err := proto.Marshal(event)
		if err != nil {
			log.Printf("marshal candidate error: %v\n", err)
			return
		}

		err = conn.WriteMessage(websocket.BinaryMessage, protobuf)
		if err != nil {
			log.Printf("conn.WriteMessage error: %v\n", err)
		}
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

		event := &pb.RoomEvent{Type: pb.RoomEvent_OFFER, From: 0, Data: []byte(offer.SDP)}
		data, err := proto.Marshal(event)
		if err != nil {
			log.Printf("marshal offer error: %v\n", err)
			return
		}

		err = conn.WriteMessage(websocket.BinaryMessage, data)
		if err != nil {
			log.Printf("conn.WriteMessage error: %v\n", err)
		}
	})

	peer.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		r.closePeer(id)
	})

	return peer, &offer, nil
}

func (r *Room) onCandidate(id int, candidate *webrtc.ICECandidateInit) {
	r.mux.Lock()
	peer, ok := r.peers[id]
	r.mux.Unlock()

	if !ok {
		// @todo
		return
	}

	err := peer.transport.AddICECandidate(*candidate)
	if err != nil {
		log.Printf("peer.AddICECandidate error: %v\n", err)
	}
}

func (r *Room) onAnswer(id int, desc webrtc.SessionDescription) {
	r.mux.Lock()
	peer, ok := r.peers[id]
	r.mux.Unlock()

	if !ok {
		// @todo
		return
	}

	err := peer.transport.SetRemoteDescription(desc)
	if err != nil {
		log.Printf("peer.SetRemoteDescription error: %v\n", err)
	}
}

func (r *Room) closePeer(id int) {
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
