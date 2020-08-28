package roomsV2

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

// @todo
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

type Transport struct {
	sync.Mutex

	id int
	pc *webrtc.PeerConnection
}

func NewTransport(id int, offer webrtc.SessionDescription) (*Transport, error) {
	// @todo this should probably be checked before calling this func
	if offer.Type != webrtc.SDPTypeOffer {
		return nil, errInvalidSDP
	}

	me := webrtc.MediaEngine{}
	err := me.PopulateFromSDP(offer)
	if err != nil {
		return nil, errSdpParseFailed
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(me))
	pc, err := api.NewPeerConnection(config)
	if err != nil {
		return nil, errConnectionInitFailed
	}

	t := &Transport{
		id: id,
		pc: pc,
	}

	pc.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver) {

	})

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {

	})

	return t, nil
}

func (t *Transport) Close() error {
	t.Lock()
	defer t.Unlock()

	return t.pc.Close()
}
