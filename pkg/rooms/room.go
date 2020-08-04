package rooms

import (
	"encoding/binary"
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

// Member is used to communicate what peers are part of the chat
type Member struct {
	ID          int      `json:"id"`
	DisplayName string   `json:"display_name"`
	Role        PeerRole `json:"role"`
	IsMuted     bool     `json:"is_muted"`
}

type Peer struct {
	Member
	connection  *webrtc.PeerConnection
	track       *webrtc.Track
	output      *webrtc.Track
	api         *webrtc.API
	dataChannel *webrtc.DataChannel
}

func (p Peer) CanSpeak() bool {
	return p.Member.Role != AUDIENCE
}

func (p Peer) GetMember() Member {
	return p.Member
}

// @todo we need to figure out how to multiplex nicely

// @todo what needs to happen is the following
//  - every connection has an input track
//  - every connection reads from their remote tracks and writes to the others if they are unmuted.

type Room struct {
	sync.RWMutex

	id   int
	name string

	peers map[int]*Peer

	disconnected chan<- bool
}

func NewRoom(id int, name string, disconnected chan bool) *Room {
	return &Room{
		id:           id,
		name:         name,
		peers:        make(map[int]*Peer),
		disconnected: disconnected,
	}
}

func (r *Room) MapPeers(fn func(int, Peer)) {
	r.RLock()
	defer r.RUnlock()

	for i, peer := range r.peers {
		fn(i, *peer)
	}
}

func (r *Room) GetID() int {
	return r.id
}

func (r *Room) GetName() string {
	return r.name
}

func (r *Room) PeerCount() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.peers)
}

func (r *Room) GetRoleForPeer(id int) PeerRole {
	r.RLock()
	defer r.RUnlock()
	return r.peers[id].Member.Role
}

func (r *Room) Join(id int, name string, offer webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
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
			//r.peerDisconnected(id)
		}
	})

	r.Lock()
	role := SPEAKER
	if len(r.peers) == 0 {
		role = OWNER
	}

	member := Member{ID: id, DisplayName: name, Role: role, IsMuted: false}
	r.peers[id] = &Peer{
		Member:     member,
		connection: peerConnection,
		output:     outputTrack,
		api:        api,
	}
	r.Unlock()

	r.setupDataChannel(id, peerConnection, channel)

	var localTrackChan = make(chan *webrtc.Track)
	// Set a handler forf when a new remote track starts, this just distributes all our packets
	// to connected gpeers
	peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		log.Println("onTrack")
		go func() {
			ticker := time.NewTicker(3 * time.Second)
			for range ticker.C {
				if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.SliceLossIndication{MediaSSRC: remoteTrack.SSRC()}}); rtcpSendErr != nil {
					// @todo
				}
			}
		}()

		data, err := json.Marshal(member)
		if err != nil {
			log.Printf("failed to encode: %s\n", err.Error())
		}

		// @todo, we should probably only set the peer here.
		go r.notify(&pb.RoomEvent{Type: pb.RoomEvent_JOINED, From: int64(id), Data: data})

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
			r.forwardPacket(id, i)
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

		r.peers[id].track = track
	}()

	log.Printf("new peer joined: %d", id)

	return peerConnection.LocalDescription(), nil
}

func (r *Room) peerDisconnected(id int) {
	r.Lock()
	defer r.Unlock()

	if r.peers[id] == nil {
		return
	}

	log.Printf("user %d left room %d", id, r.id)

	role := r.peers[id].Role

	r.peers[id].connection.Close()
	delete(r.peers, id)
	r.disconnected <- true

	go r.notify(&pb.RoomEvent{Type: pb.RoomEvent_LEFT, From: int64(id)})

	if role == OWNER {
		go r.electOwner()
	}

}

func (r *Room) electOwner() {
	r.Lock()
	defer r.Unlock()

	speaker := first(r.peers, func(peer *Peer) bool {
		return peer.Role == SPEAKER
	})

	if speaker != 0 {
		r.peers[speaker].Role = OWNER
	} else {
		for k, v := range r.peers {
			v.Role = OWNER
			speaker = k
			break
		}
	}

	go r.notify(&pb.RoomEvent{Type: pb.RoomEvent_CHANGED_OWNER, Data: intToBytes(speaker)})
}

func (r *Room) setupDataChannel(id int, peer *webrtc.PeerConnection, channel *webrtc.DataChannel) {
	handler := func(msg webrtc.DataChannelMessage) {
		cmd := &pb.RoomCommand{}
		err := proto.Unmarshal(msg.Data, cmd)
		if err != nil {
			log.Printf("failed to decode data channel message: %s\n", err.Error())
		}

		r.onCommand(id, cmd)
	}

	channel.OnClose(func() {
		r.peerDisconnected(id)
	})

	channel.OnMessage(handler)

	peer.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			r.Lock()
			defer r.Unlock()

			r.peers[id].dataChannel = d
		})

		d.OnClose(func() {
			r.peerDisconnected(id)
		})

		d.OnMessage(handler)
	})
}

func (r *Room) forwardPacket(from int, packet *rtp.Packet) {
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

func (r *Room) onCommand(from int, command *pb.RoomCommand) {
	switch command.Type {
	case pb.RoomCommand_ADD_SPEAKER:
		r.onAddSpeaker(from, command.Data)
	case pb.RoomCommand_REMOVE_SPEAKER:
		r.onRemoveSpeaker(from, command.Data)
	case pb.RoomCommand_MUTE_SPEAKER:
		r.onMuteSpeaker(from)
	case pb.RoomCommand_UNMUTE_SPEAKER:
		r.onUnmuteSpeaker(from)
	}
}

func (r *Room) onAddSpeaker(from int, peer []byte) {
	r.Lock()
	defer r.Unlock()

	if r.peers[from].Role != OWNER {
		return
	}
	r.peers[bytesToInt(peer)].Role = SPEAKER

	go r.notify(&pb.RoomEvent{Type: pb.RoomEvent_ADDED_SPEAKER, From: int64(from), Data: peer})
}

func (r *Room) onRemoveSpeaker(from int, peer []byte) {
	r.Lock()
	defer r.Unlock()

	if r.peers[from].Role != OWNER {
		return
	}
	r.peers[bytesToInt(peer)].Role = AUDIENCE

	go r.notify(&pb.RoomEvent{Type: pb.RoomEvent_REMOVED_SPEAKER, From: int64(from), Data: peer})
}

func (r *Room) onMuteSpeaker(from int) {
	r.RLock()
	peer := r.peers[from]
	r.RUnlock()

	if peer.IsMuted {
		return
	}

	r.Lock()
	peer.IsMuted = true
	r.Unlock()

	go r.notify(&pb.RoomEvent{Type: pb.RoomEvent_MUTED_SPEAKER, From: int64(from)})
	log.Printf("user %s muted", from)
}

func (r *Room) onUnmuteSpeaker(from int) {
	r.RLock()
	peer := r.peers[from]
	r.RUnlock()

	if !peer.IsMuted {
		return
	}
	r.Lock()
	peer.IsMuted = false
	r.Unlock()

	go r.notify(&pb.RoomEvent{Type: pb.RoomEvent_UNMUTED_SPEAKER, From: int64(from)})
	log.Printf("user %d unmuted", from)
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
		if int64(id) == event.From {
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

func first(peers map[int]*Peer, fn func(*Peer) bool) int {
	for i, peer := range peers {
		if fn(peer) {
			return i
		}
	}

	return 0
}

func intToBytes(val int) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(val))
	return bytes
}

func bytesToInt(val []byte) int {
	return int(binary.LittleEndian.Uint64(val))
}
