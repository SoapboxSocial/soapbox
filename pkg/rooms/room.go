package rooms

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
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

type Peer struct {
	isOwner    bool
	isMuted    bool
	connection *webrtc.PeerConnection
	track      *webrtc.Track
	output     *webrtc.Track
	api        *webrtc.API
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

func (r *Room) GetID() int {
	return r.id
}

func (r *Room) PeerCount() int {
	r.RLock()
	defer r.RUnlock()
	return len(r.peers)
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
		// @todo, this does not seem completely safe
		// @todo disconnected here is certainly not reliable
		if state == webrtc.PeerConnectionStateClosed || state == webrtc.PeerConnectionStateFailed /* || state == webrtc.PeerConnectionStateDisconnected */ {
			//	// @todo this seems like it could be buggy
			delete(r.peers, addr)
			r.disconnected <- true
		}
	})

	var localTrackChan = make(chan *webrtc.Track)
	// Set a handler forf when a new remote track starts, this just distributes all our packets
	// to connected peers
	peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		go func() {
			ticker := time.NewTicker(3 * time.Second)
			for range ticker.C {
				rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{
					&rtcp.SliceLossIndication{MediaSSRC: remoteTrack.SSRC()},
				})

				if rtcpSendErr != nil {
					// @todo
				}
			}
		}()

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

		r.handleTrack(peerConnection, remoteTrack)
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

	r.peers[addr] = &Peer{
		isOwner:    len(r.peers) == 0,
		isMuted:    false,
		connection: peerConnection,
		output:     outputTrack,
		api:        api,
	}

	// @todo this needs be elsewhere
	go func() {
		track := <-localTrackChan

		r.peers[addr].track = track
	}()

	return peerConnection.LocalDescription(), nil
}

func (r *Room) handleTrack(peer *webrtc.PeerConnection, remote *webrtc.Track) {
	for {
		i, readErr := remote.ReadRTP()
		if readErr != nil {
			log.Printf("failed to read from remote track: %s", readErr.Error())
			peer.Close()
			return
		}

		r.RLock()
		for _, p := range r.peers {
			if p.output == nil {
				continue
			}

			if p.connection == peer {
				continue
			}

			err := p.output.WriteRTP(i)
			if err != nil && err != io.ErrClosedPipe {
				log.Printf("failed to write to track: %s", err.Error())
				peer.Close()
				return
			}
		}

		r.RUnlock()
	}
}
