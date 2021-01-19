package rooms

import (
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

func NewServer(
	sfu *sfu.SFU,
	sm *sessions.SessionManager,
	ub *users.UserBackend,
	queue *pubsub.Queue,
	currentRoom *CurrentRoomBackend,
	groups *groups.Backend,
) *Server {
	return &Server{
		sfu:         sfu,
		sm:          sm,
		ub:          ub,
		queue:       queue,
		currentRoom: currentRoom,
		groups:      groups,
		rooms:       make(map[string]*Room),
	}
}

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

	user, err := s.userForSession(session)
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
			SDP:  join.Description.Sdp,
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
						Sdp:  answer.SDP,
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

		id := internal.GenerateRoomID()
		name := internal.TrimRoomNameToLimit(create.Name)

		//var group *groups.Group
		// @TODO
		//if payload.Create.GetGroup() != 0 {
		//	group, err = s.getGroup(user.ID, int(payload.Create.GetGroup()))
		//	if err != nil {
		//		return status.Errorf(codes.Internal, "group error %s", err)
		//	}
		//}

		room = NewRoom(id, name) // @TODO THE REST

		description := webrtc.SessionDescription{
			Type: newSDPType(create.Description.Type),
			SDP:  create.Description.Sdp,
		}

		answer, err := setup(peer, id, stream, description)
		if err != nil {
			return err
		}

		err = stream.Send(&pb.SignalReply{
			Id: in.Id,
			Payload: &pb.SignalReply_Create{
				Create: &pb.CreateReply{
					Id: id,
					Description: &pb.SessionDescription{
						Type: answer.Type.String(),
						Sdp:  answer.SDP,
					},
				},
			},
		})

		if err != nil {
			log.Printf("error sending create response %s", err)
			return status.Errorf(codes.Internal, "create error %s", err)
		}

		s.mux.Lock()
		s.rooms[id] = room
		s.mux.Unlock()

	default:
		return status.Error(codes.FailedPrecondition, "invalid message")
	}

	errChan := make(chan error)
	go func() {
		err := room.Handle(user.ID, peer) // @TODO
		errChan <- err
	}()

	for {
		select {
		case err := <-errChan:
			log.Printf("handle err %v", err)
			_ = peer.Close()
			return nil
		default:
			in, err := receive(peer, stream)
			if err != nil {
				return err
			}

			err = s.handle(peer, stream, in)
			if err != nil {
				return err
			}
		}
	}
}

func (s *Server) handle(peer *sfu.Peer, stream pb.SFU_SignalServer, in *pb.SignalRequest) error {
	switch in.Payload.(type) {
	case *pb.SignalRequest_Description:
		payload := in.GetDescription()
		sdp := webrtc.SessionDescription{
			Type: newSDPType(payload.Type),
			SDP:  payload.Sdp,
		}

		if sdp.Type == webrtc.SDPTypeOffer {
			answer, err := peer.Answer(sdp)
			if err != nil {
				switch err {
				case sfu.ErrNoTransportEstablished:
					fallthrough
				case sfu.ErrOfferIgnored:
					err = stream.Send(&pb.SignalReply{
						Payload: &pb.SignalReply_Error{
							Error: fmt.Errorf("negotiate answer error: %w", err).Error(),
						},
					})
					if err != nil {
						log.Printf("grpc send error %v\n", err)
						return status.Errorf(codes.Internal, err.Error())
					}
					return nil
				default:
					return status.Errorf(codes.Unknown, fmt.Sprintf("negotiate error: %v", err))
				}
			}

			err = stream.Send(&pb.SignalReply{
				Id: in.Id,
				Payload: &pb.SignalReply_Description{
					Description: &pb.SessionDescription{
						Type: answer.Type.String(),
						Sdp:  answer.SDP,
					},
				},
			})

			if err != nil {
				return status.Errorf(codes.Internal, fmt.Sprintf("negotiate error: %v", err))
			}

		} else if sdp.Type == webrtc.SDPTypeAnswer {
			err := peer.SetRemoteDescription(sdp)
			if err != nil {
				switch err {
				case sfu.ErrNoTransportEstablished:
					err = stream.Send(&pb.SignalReply{
						Payload: &pb.SignalReply_Error{
							Error: fmt.Errorf("set remote description error: %w", err).Error(),
						},
					})
					if err != nil {
						log.Printf("grpc send error %v\n", err)
						return status.Errorf(codes.Internal, err.Error())
					}
				default:
					return status.Errorf(codes.Unknown, err.Error())
				}
			}
		}
	case *pb.SignalRequest_Trickle:
		payload := in.GetTrickle()

		midLine := uint16(payload.IceCandidate.SdpMLineIndex)
		candidate := webrtc.ICECandidateInit{
			Candidate: payload.IceCandidate.Candidate,
			SDPMid: &payload.IceCandidate.SdpMid,
			SDPMLineIndex: &midLine,
		}

		err := peer.Trickle(candidate, int(payload.Target))
		if err != nil {
			switch err {
			case sfu.ErrNoTransportEstablished:
				log.Print("peer hasn't joined")
				err = stream.Send(&pb.SignalReply{
					Payload: &pb.SignalReply_Error{
						Error: fmt.Errorf("trickle error:  %w", err).Error(),
					},
				})
				if err != nil {
					log.Printf("grpc send error %v\n", err)
					return status.Errorf(codes.Internal, err.Error())
				}
			default:
				return status.Errorf(codes.Unknown, fmt.Sprintf("negotiate error: %v", err))
			}
		}
	}

	return nil
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
			Sdp:  o.SDP,
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
			log.Printf("OnICEConnectionStateChange error %s", err)
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

func (s *Server) userForSession(session string) (*users.User, error) {
	id, err := s.sm.GetUserIDForSession(session)
	if err != nil {
		return nil, err
	}

	u, err := s.ub.FindByID(id)
	if err != nil {
		return nil, err
	}

	return u, nil
}