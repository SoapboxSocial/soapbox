package rooms

import (
	"encoding/json"
	"log"

	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
)

type Server struct {
	room *RoomLegacy
	sfu  *sfu.SFU
	sm   *sessions.SessionManager
}

func NewServer(sfu *sfu.SFU) *Server {
	return &Server{
		sfu: sfu,
	}
}

func (s *Server) Signal(stream pb.RoomService_SignalServer) error {
	var room *Room

	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}

		switch payload := in.Payload.(type) {
		case *pb.SignalRequest_Join:
			break
		case *pb.SignalRequest_Create:
			break
		case *pb.SignalRequest_Trickle:
		default:
			break
		}
	}

	//	switch payload := in.Payload.(type) {
	//	case *pb.SignalRequest_Join:
	//		if peer != nil {
	//			// @todo
	//			return status.Errorf(codes.FailedPrecondition, "peer already exists")
	//		}
	//
	//		// @todo pass some channel to listen to messages
	//		// @todo Join ID needs be peers ID.
	//		p, offer, err := s.room.Join(int(payload.Join.RoomLegacy))
	//		if err != nil {
	//			return status.Errorf(codes.FailedPrecondition, "join error: %s", err)
	//		}
	//
	//		peer = p
	//
	//		err = stream.Send(&pb.SignalReply{
	//			Payload: &pb.SignalReply_Join{
	//				Join: &pb.JoinReply{
	//					Answer: &pb.SessionDescription{
	//						Type: offer.Type.String(),
	//						Sdp:  []byte(offer.SDP),
	//					},
	//				},
	//			},
	//		})
	//
	//		if err != nil {
	//			return status.Errorf(codes.Internal, "join error %s", err)
	//		}
	//	case *pb.SignalRequest_Negotiate:
	//		if peer == nil {
	//			return status.Errorf(codes.FailedPrecondition, "%s", errNoPeer)
	//		}
	//
	//		if payload.Negotiate.Type != webrtc.SDPTypeAnswer.String() {
	//			return status.Errorf(codes.FailedPrecondition, "invalid negotiation %s", payload.Negotiate.Type)
	//		}
	//
	//		err = peer.SetRemoteDescription(webrtc.SessionDescription{
	//			Type: webrtc.SDPTypeAnswer,
	//			SDP:  string(payload.Negotiate.Sdp),
	//		})
	//
	//		if err != nil {
	//			return status.Errorf(codes.Internal, "%s", err)
	//		}
	//	case *pb.SignalRequest_Trickle:
	//		if peer == nil {
	//			return status.Errorf(codes.FailedPrecondition, "%s", errNoPeer)
	//		}
	//
	//		var candidate webrtc.ICECandidateInit
	//		err := json.Unmarshal([]byte(payload.Trickle.Init), &candidate)
	//		if err != nil {
	//			log.Printf("error parsing ice candidate: %v", err)
	//		}
	//
	//		if err := peer.AddICECandidate(candidate); err != nil {
	//			return status.Errorf(codes.Internal, "error adding ice candidate")
	//		}
	//	}
	//}
}
