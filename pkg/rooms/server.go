package rooms

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
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

const MAX_PEERS = 16

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

// @TODO WEBRTC
func (s *Server) Signal(stream pb.SFU_SignalServer) error {
	peer := sfu.NewPeer(s.sfu)

	in, err := receive(peer, stream)
	if err != nil {
		return err
	}

	if in == nil {
		return nil
	}

	session, err := internal.SessionID(stream.Context())
	if err != nil {
		_ = peer.Close()
	}

	var room *Room

	switch in.Payload.(type) {
	case *pb.SignalRequest_Join:
		join := in.GetJoin()
		if join == nil {
			return status.Errorf(codes.Internal, "something went wrong")
		}

		s.mux.RLock()
		r, ok := s.rooms[join.Room]
		s.mux.RUnlock()

		if !ok {
			return status.Errorf(codes.Internal, "join error room closed")
		}

		if r.PeerCount() >= MAX_PEERS {
			return status.Errorf(codes.Internal, "join error room full")
		}

		//if !s.canJoin(user.ID, r) {
		//	return status.Errorf(codes.Internal, "user not invited")
		//}

		description := webrtc.SessionDescription{
			Type: newSDPType(join.Description.Type),
			SDP:  string(join.Description.Sdp),
		}

		answer, err := setup(peer, join.Room, stream, description)
		if err != nil {
			return err
		}

		err = stream.Send(&pb.SignalReply{
			Id: in.Id,
			Payload: &pb.SignalReply_Join{
				Join: &pb.JoinReply{
					//Room: r.ToProtoForPeer(),
					Description: &pb.SessionDescription{
						Type: answer.Type.String(),
						Sdp:  []byte(answer.SDP),
					},
				},
			},
		})

		if err != nil {
			log.Printf("error sending join response %s", err)
			return status.Errorf(codes.Internal, "join error %s", err)
		}

		room = r
	case *pb.SignalRequest_Create:
		create := in.GetCreate()
		if create == nil {
			return status.Errorf(codes.Internal, "something went wrong")
		}
	default:
		return status.Error(codes.FailedPrecondition, "invalid message")
	}

	// @TODO Connect to room?

	errChan := make(chan error)
	go func() {
		err := room.Handle(id, peer) // @TODO
		errChan <- err
	}()

	for {
		select {
		case <-errChan:
			// @TODO
			return nil
		default:
		}
	}
}

func receive(peer *sfu.Peer, stream pb.SFU_SignalServer) (*pb.SignalRequest, error) {
	in, err := stream.Recv()
	if err != nil {
		_ = peer.Close()

		if err == io.EOF {
			return nil, nil
		}

		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.Canceled {
			return nil, nil
		}

		log.Printf("signal error %v %v", errStatus.Message(), errStatus.Code())
		return nil, err
	}

	return in, nil
}

// setup properly sets up a peer to communicate with the SFU.
func setup(peer *sfu.Peer, room string, stream pb.SFU_SignalServer, description webrtc.SessionDescription) (*webrtc.SessionDescription, error) {

	// Notify user of new ice candidate
	peer.OnIceCandidate = func(candidate *webrtc.ICECandidateInit, target int) {
		candidateProto := &pb.ICECandidate{
			Candidate: candidate.Candidate,
		}

		if candidate.SDPMid != nil {
			candidateProto.SdpMid = *candidate.SDPMid
		}

		if candidate.SDPMLineIndex != nil {
			candidateProto.SdpMLineIndex = int64(*candidate.SDPMLineIndex)
		}

		if candidate.UsernameFragment != nil {
			candidateProto.UsernameFragment = *candidate.UsernameFragment
		}

		err := stream.Send(&pb.SignalReply{
			Payload: &pb.SignalReply_Trickle{
				Trickle: &pb.Trickle{
					IceCandidate: candidateProto,
					Target:       pb.Trickle_Target(target),
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
			Sdp:  []byte(o.SDP),
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
				return nil, status.Errorf(codes.Internal, err.Error())
			}
		default:
			return nil, status.Errorf(codes.Unknown, err.Error())
		}
	}

	return answer, nil
}

func newSDPType(raw string) webrtc.SDPType {
	val := strings.ToLower(raw)

	switch val {
	case "offer":
		return webrtc.SDPTypeOffer
	case "pranswer":
		return webrtc.SDPTypePranswer
	case "answer":
		return webrtc.SDPTypeAnswer
	case "rollback":
		return webrtc.SDPTypeRollback
	default:
		return webrtc.SDPType(webrtc.Unknown)
	}
}
