package rooms

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"

	"github.com/ephemeral-networks/voicely/pkg/pb"
)

var config = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{
				"stun:stun.l.google.com:19302",
				"stun:stun1.l.google.com:19302",
				"stun:stun2.l.google.com:19302",
				"stun:stun3.l.google.com:19302",
				"stun:stun4.l.google.com:19302",
			},
		},
	},
}

type PeerRole string

const (
	OWNER    PeerRole = "owner"
	SPEAKER  PeerRole = "speaker"
	AUDIENCE PeerRole = "audience"
)

type Peer struct {
	role        PeerRole
	isMuted     bool
	connection  *webrtc.PeerConnection
	track       *webrtc.Track
	output      *webrtc.Track
	api         *webrtc.API
	dataChannel *webrtc.DataChannel
}

func (p Peer) CanSpeak() bool {
	return p.role != AUDIENCE
}

func (p Peer) Role() PeerRole {
	return p.role
}

// @todo we need to figure out how to multiplex nicely

// @todo what needs to happen is the following
//  - every connection has an input track
//  - every connection reads from their remote tracks and writes to the others if they are unmuted.

type Room struct {
	sync.RWMutex

	id int

	peers map[string]*Peer

	disconnected chan<- bool
}

func NewRoom(id int, disconnected chan bool) *Room {
	return &Room{
		id:           id,
		peers:        make(map[string]*Peer),
		disconnected: disconnected,
	}
}

func (r *Room) MapPeers(fn func(string, Peer)) {
	r.RLock()
	defer r.RUnlock()

	for i, peer := range r.peers {
		fn(i, *peer)
	}
}

func (r *Room) GetID() int {
	return r.id
}

func (r *Room) PeerCount() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.peers)
}

func (r *Room) GetRoleForPeer(id string) PeerRole {
	r.RLock()
	defer r.RUnlock()
	return r.peers[id].role
}

func (r *Room) Join(addr string, offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	mediaEngine := webrtc.MediaEngine{}
	err := mediaEngine.PopulateFromSDP(offer)
	if err != nil {
		return nil, err
	}

	// Create the API object with the MediaEngine
	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	// Create a new RTCPeerConnection
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	codecs := mediaEngine.GetCodecsByKind(webrtc.RTPCodecTypeAudio)

	outputTrack, err := peerConnection.NewTrack(codecs[0].PayloadType, rand.Uint32(), "audio0", "pion")
	if err != nil {
		return nil, err
	}

	channel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		return nil, err
	}

	peerConnection.AddTrack(outputTrack)

	// Allow us to receive 1 video track
	if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		return nil, err
	}

	peerConnection.OnSignalingStateChange(func(state webrtc.SignalingState) {
		// @todo
	})

	peerConnection.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		fmt.Println(state)
		// @todo I think this should work.
		// @todo, this does not seem completely safe
		// @todo disconnected here is certainly not reliable
		if state == webrtc.PeerConnectionStateClosed || state == webrtc.PeerConnectionStateFailed /* || state == webrtc.PeerConnectionStateDisconnected */ {
			//	// @todo this seems like it could be buggy
			r.peerDisconnected(addr)
		}
	})

	r.Lock()
	role := SPEAKER
	if len(r.peers) == 0 {
		role = OWNER
	}

	r.peers[addr] = &Peer{
		role:       role,
		isMuted:    false,
		connection: peerConnection,
		output:     outputTrack,
		api:        api,
	}
	r.Unlock()

	r.setupDataChannel(addr, peerConnection, channel)

	var localTrackChan = make(chan *webrtc.Track)
	// Set a handler forf when a new remote track starts, this just distributes all our packets
	// to connected peers
	peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		go func() {
			ticker := time.NewTicker(3 * time.Second)
			for range ticker.C {
				if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.SliceLossIndication{MediaSSRC: remoteTrack.SSRC()}}); rtcpSendErr != nil {
					// @todo
				}
			}
		}()

		data, err := json.Marshal(struct {
			ID   string `json:"id"`
			Role string `json:"role"`
		}{ID: addr, Role: string(role)})
		if err != nil {
			log.Printf("failed to encode: %s\n", err.Error())
		}

		// @todo handle join here, sending a member encoded as json
		go r.notify(&pb.RoomEvent{Type: pb.RoomEvent_JOINED, From: addr, Data: data})

		// Create a local track, all our SFU clients will be fed via this track
		localTrack, newTrackErr := peerConnection.NewTrack(
			remoteTrack.PayloadType(),
			remoteTrack.SSRC(),
			"audio0", "pion",
		)

		if newTrackErr != nil {
			log.Printf("failed to create new track: %s", newTrackErr.Error())
			peerConnection.Close()
			return
		}
		localTrackChan <- localTrack

		for {
			i, readErr := remoteTrack.ReadRTP()
			if readErr != nil {
				log.Printf("failed to read from remote track: %s", readErr.Error())
				peerConnection.Close()
				return
			}

			// @todo maybe if we push into a channel here, we can read from it, and see if concurrency issue still occurs.
			r.forwardPacket(addr, i)
		}
	})

	// Set the remote SessionDescription
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		return nil, err
	}

	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}

	// Create channel that is blocked until ICE Gathering is complete
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		return nil, err
	}
	// Block until ICE Gathering is complete, disabling trickle ICE
	// we do this because we only can exchange one signaling message
	// in a production application you should exchange ICE Candidates via OnICECandidate
	<-gatherComplete

	// @todo this needs be elsewhere
	go func() {
		track := <-localTrackChan

		r.peers[addr].track = track
	}()

	log.Printf("new peer joined: %s", addr)

	return peerConnection.LocalDescription(), nil
}

func (r *Room) peerDisconnected(addr string) {
	r.Lock()
	defer r.Unlock()

	if r.peers[addr] == nil {
		return
	}

	role := r.peers[addr].role

	r.peers[addr].connection.Close()
	delete(r.peers, addr)
	r.disconnected <- true

	go r.notify(&pb.RoomEvent{Type: pb.RoomEvent_LEFT, From: addr})

	if role == OWNER {
		go r.electOwner()
	}

}

func (r *Room) electOwner() {
	r.Lock()
	defer r.Unlock()

	speaker := first(r.peers, func(peer *Peer) bool {
		return peer.role == SPEAKER
	})

	if speaker != "" {
		r.peers[speaker].role = OWNER
	} else {
		for k, v := range r.peers {
			v.role = OWNER
			speaker = k
			break
		}
	}

	go r.notify(&pb.RoomEvent{Type: pb.RoomEvent_CHANGED_OWNER, Data: []byte(speaker)})
}

func (r *Room) setupDataChannel(addr string, peer *webrtc.PeerConnection, channel *webrtc.DataChannel) {
	handler := func(msg webrtc.DataChannelMessage) {
		cmd := &pb.RoomCommand{}
		err := proto.Unmarshal(msg.Data, cmd)
		if err != nil {
			log.Printf("failed to decode data channel message: %s\n", err.Error())
		}

		r.onCommand(addr, cmd)
	}

	channel.OnClose(func() {
		r.peerDisconnected(addr)
	})

	channel.OnMessage(handler)

	peer.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			r.Lock()
			defer r.Unlock()

			r.peers[addr].dataChannel = d
		})

		d.OnClose(func() {
			r.peerDisconnected(addr)
		})

		d.OnMessage(handler)
	})
}

func (r *Room) forwardPacket(from string, packet *rtp.Packet) {
	r.RLock()
	defer r.RUnlock()

	if !r.peers[from].CanSpeak() {
		return
	}

	for id, p := range r.peers {
		if p.output == nil {
			continue
		}

		if id == from {
			continue
		}

		err := p.output.WriteRTP(packet)
		if err != nil && err != io.ErrClosedPipe {
			log.Printf("failed to write to track: %s", err.Error())
		}
	}
}

func (r *Room) onCommand(from string, command *pb.RoomCommand) {
	switch command.Type {
	case pb.RoomCommand_ADD_SPEAKER:
		r.onAddSpeaker(from, string(command.Data))
	case pb.RoomCommand_REMOVE_SPEAKER:
		r.onRemoveSpeaker(from, string(command.Data))
	}
}

func (r *Room) onAddSpeaker(from, peer string) {
	r.Lock()
	defer r.Unlock()

	if r.peers[from].role != OWNER {
		return
	}
	r.peers[peer].role = SPEAKER

	go r.notify(&pb.RoomEvent{Type: pb.RoomEvent_ADDED_SPEAKER, From: from, Data: []byte(peer)})
}

func (r *Room) onRemoveSpeaker(from, peer string) {
	r.Lock()
	defer r.Unlock()

	if r.peers[from].role != OWNER {
		return
	}
	r.peers[peer].role = AUDIENCE

	go r.notify(&pb.RoomEvent{Type: pb.RoomEvent_REMOVED_SPEAKER, From: from, Data: []byte(peer)})
}

func (r *Room) notify(event *pb.RoomEvent) {
	r.RLock()
	defer r.RUnlock()

	data, err := proto.Marshal(event)
	if err != nil {
		// @todo
		return
	}

	for id, p := range r.peers {
		if id == event.From {
			continue
		}

		if p.dataChannel == nil {
			continue
		}

		err := p.dataChannel.Send(data)
		if err != nil {
			// @todo
			log.Printf("failed to write to data channel: %s\n", err.Error())
		}
	}
}

func first(peers map[string]*Peer, fn func(*Peer) bool) string {
	for i, peer := range peers {
		if fn(peer) {
			return i
		}
	}

	return ""
}
