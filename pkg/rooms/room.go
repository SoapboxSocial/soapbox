package rooms

import (
	"encoding/json"
	"log"
	"sync"

	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type Room struct {
	mux sync.RWMutex
}

func (r *Room) Handle(stream pb.RoomService_SignalServer, peer *sfu.WebRTCTransport) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}

		switch payload := in.Payload.(type) {
		case *pb.SignalRequest_Negotiate:
			if payload.Negotiate.Type != "answer" {
				// @todo
				continue
			}

			err = peer.SetRemoteDescription(webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  string(payload.Negotiate.Sdp),
			})

			if err != nil {
				return status.Errorf(codes.Internal, "%s", err)
			}
		case *pb.SignalRequest_Trickle:
			log.Print("yay")
			var candidate webrtc.ICECandidateInit
			err := json.Unmarshal([]byte(payload.Trickle.Init), &candidate)
			if err != nil {
				log.Printf("error parsing ice candidate: %v", err)
			}

			err = peer.AddICECandidate(candidate)
			if err != nil {
				return status.Errorf(codes.Internal, "error adding ice candidate")
			}
		}
	}
}
