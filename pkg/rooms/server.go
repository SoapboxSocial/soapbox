package rooms

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/soapboxsocial/soapbox/pkg/groups"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms/internal"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Server struct {
	mux sync.RWMutex

	sfu    *sfu.SFU
	sm     *sessions.SessionManager
	ub     *users.UserBackend
	groups *groups.Backend
	queue  *pubsub.Queue

	currentRoom *CurrentRoomBackend

	rooms map[string]*Room
}

func (s *Server) Signal(stream pb.SFU_SignalServer) error {
	peer := sfu.NewPeer(s.sfu)

	in, err := stream.Recv()
	if err != nil {
		_ = peer.Close()

		if err == io.EOF {
			return nil
		}

		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.Canceled {
			return nil
		}

		log.Printf("signal error %v %v", errStatus.Message(), errStatus.Code())
		return err
	}

	id, err := internal.SessionID(stream.Context())
	if err != nil {
		_ = peer.Close()
	}

	switch in.Payload.(type) {
	case *pb.SignalRequest_Join:
		break
	case *pb.SignalRequest_Create:
		break
	default:
		return status.Error(codes.FailedPrecondition, "invalid message")
	}

	// @TODO Connect to room?

	for {
		// @TODO READ OTHER PACKETS
	}
}

// setup properly sets up a peer to communicate with the SFU.
func setup(peer *sfu.Peer, messageId, room string, stream pb.SFU_SignalServer, description webrtc.SessionDescription) error {

	// Notify user of new ice candidate
	peer.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
		bytes, err := json.Marshal(candidate)
		if err != nil {
			log.Printf("OnIceCandidate error %s", err)
		}
		err = stream.Send(&pb.SignalReply{
			Payload: &pb.SignalReply_Trickle{
				Trickle: &pb.Trickle{
					Init:   string(bytes),
					Target: pb.Trickle_Target(target),
				},
			},
		})
		if err != nil {
			log.Printf("OnIceCandidate send error %v ", err)
		}
	}

	// Notify user of new offer
	peer.OnOffer = func(o *webrtc.SessionDescription) {
		description := &pb.SessionDescription{
			Type: o.Type.String(),
			Sdp: []byte(o.SDP),
		}

		err := stream.Send(&pb.SignalReply{
			Payload: &pb.SignalReply_Description{
				Description: description,
			},
		})

		if err != nil {
			log.Printf("negotiation error %s", err)
		}
	}

	peer.OnICEConnectionStateChange = func(c webrtc.ICEConnectionState) {
		err := stream.Send(&pb.SignalReply{
			Payload: &pb.SignalReply_IceConnectionState{
				IceConnectionState: c.String(),
			},
		})

		if err != nil {
			log.Printf("oniceconnectionstatechange error %s", err)
		}
	}

	answer, err := peer.Join(room, description)
	if err != nil {
		switch err {
		case sfu.ErrTransportExists:
			fallthrough
		case sfu.ErrOfferIgnored:
			err = stream.Send(&pb.SignalReply{
				Payload: &pb.SignalReply_Error{
					Error: fmt.Errorf("join error: %w", err).Error(),
				},
			})
			if err != nil {
				log.Printf("grpc send error %v ", err)
				return status.Errorf(codes.Internal, err.Error())
			}
		default:
			return status.Errorf(codes.Unknown, err.Error())
		}
	}

	// send answer
	err = stream.Send(&pb.SignalReply{
		Id: messageId,
		Payload: &pb.SignalReply_Join{
			Join: &pb.JoinReply{
				Description: &pb.SessionDescription{
					Type: answer.Type.String(),
					Sdp: []byte(answer.SDP),
				},
			},
		},
	})

	if err != nil {
		log.Printf("error sending join response %s", err)
		return status.Errorf(codes.Internal, "join error %s", err)
	}

	return nil
}