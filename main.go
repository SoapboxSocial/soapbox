package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/rtcp"
	"github.com/pion/webrtc"
)

const rtcpPLIInterval = time.Second * 3

var upgrader = websocket.Upgrader{} // use default options
type Msg struct {
	messageType int
	p []byte
}

type Payload struct {
	Payload map[string]interface{} `json:"payload"`
	Type string `json:"type"`
}

type ConnectionRequest struct {
	offer webrtc.SessionDescription
	remote net.Addr
}

type Peer struct {
	isMuted    bool
	connection *webrtc.PeerConnection
	track      *webrtc.Track
	api *webrtc.API
}

type Room struct {
	peers []*Peer
}

var conns = make(map[net.Addr]chan Msg)
var requests = make(chan ConnectionRequest)

func main() {
	log.SetFlags(0)
	http.HandleFunc("/", connect)
	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	server()
}

func server() {
	var localTrack *webrtc.Track

	for {
		conn := <-requests

		mediaEngine := webrtc.MediaEngine{}
		err := mediaEngine.PopulateFromSDP(conn.offer)
		if err != nil {
			panic(err)
		}

		// Create the API object with the MediaEngine
		api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

		peerConnectionConfig := webrtc.Configuration{
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

		// Create a new RTCPeerConnection
		peerConnection, err := api.NewPeerConnection(peerConnectionConfig)
		if err != nil {
			panic(err)
		}

		if localTrack != nil {
			peerConnection.AddTrack(localTrack) // @todo get error
		}

		// Allow us to receive 1 video track
		if _, err = peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
			panic(err)
		}

		peerConnection.OnSignalingStateChange(func(state webrtc.SignalingState) {
			fmt.Println(state)
		})

		var localTrackChan chan *webrtc.Track
		if localTrack == nil {
			localTrackChan = make(chan *webrtc.Track)
			// Set a handler for when a new remote track starts, this just distributes all our packets
			// to connected peers
			peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
				// Send a PLI on an interval so that the publisher is pushing a keyframe every rtcpPLIInterval
				// This can be less wasteful by processing incoming RTCP events, then we would emit a NACK/PLI when a viewer requests it
				go func() {
					ticker := time.NewTicker(rtcpPLIInterval)
					for range ticker.C {
						if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.SliceLossIndication{MediaSSRC: remoteTrack.SSRC()}}); rtcpSendErr != nil {
							fmt.Println(rtcpSendErr)
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
					panic(newTrackErr)
				}
				localTrackChan <- localTrack

				rtpBuf := make([]byte, 1400)
				for {
					i, readErr := remoteTrack.Read(rtpBuf)
					if readErr != nil {
						panic(readErr)
					}

					// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
					if _, err = localTrack.Write(rtpBuf[:i]); err != nil && err != io.ErrClosedPipe {
						panic(err)
					}
				}
			})
		}

		// Set the remote SessionDescription
		err = peerConnection.SetRemoteDescription(conn.offer)
		if err != nil {
			panic(err)
		}

		// Create answer
		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			panic(err)
		}

		// Create channel that is blocked until ICE Gathering is complete
		gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

		// Sets the LocalDescription, and starts our UDP listeners
		err = peerConnection.SetLocalDescription(answer)
		if err != nil {
			panic(err)
		}

		// Block until ICE Gathering is complete, disabling trickle ICE
		// we do this because we only can exchange one signaling message
		// in a production application you should exchange ICE Candidates via OnICECandidate
		<-gatherComplete

		// Get the LocalDescription and take it to base64 so we can paste in browser
		pack := Payload{
			Type:    "SessionDescription",
			Payload: make(map[string]interface{}),
		}

		pack.Payload["sdp"] = peerConnection.LocalDescription().SDP
		pack.Payload["type"] = strings.ToLower(peerConnection.LocalDescription().Type.String())

		send, err := json.Marshal(pack)
		if err != nil {
			panic(err)
		}

		conns[conn.remote] <- Msg{messageType: 2, p: send}

		if localTrack == nil && localTrackChan != nil {
			localTrack = <-localTrackChan
		}
	}
}

func connect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	out := make(chan Msg)

	conns[c.RemoteAddr()] = out

	defer func() {
		delete(conns, c.RemoteAddr())
		c.Close()
		fmt.Println("closing")
	}()

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			payload := &Payload{}
			err = json.Unmarshal(message, payload)
			if err != nil {
				fmt.Println(err)
				return
			}

			if payload.Type == "SessionDescription" {
				p := webrtc.SessionDescription{
					Type: getType(payload.Payload["type"].(string)),
					SDP: payload.Payload["sdp"].(string),
				}

				requests <- ConnectionRequest{p, c.RemoteAddr()}
			}
		}
	}()

	for {
		msg := <- out
		err := c.WriteMessage(msg.messageType, msg.p)
		if err != nil {
			return
		}
	}
}

func getType(t string) webrtc.SDPType {
	switch t {
	case "offer":
		return webrtc.SDPTypeOffer
	case "prAnswer":
		return webrtc.SDPTypePranswer
	case "answer":
		return webrtc.SDPTypeAnswer
	}

	panic("fuck")
}