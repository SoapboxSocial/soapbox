package rooms

import (
	"fmt"
	"io"
	"log"
	"strings"

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
	sfu    *sfu.SFU
	sm     *sessions.SessionManager
	ub     *users.UserBackend
	groups *groups.Backend
	queue  *pubsub.Queue

	currentRoom *CurrentRoomBackend

	repository *Repository
}

func NewServer(
	sfu *sfu.SFU,
	sm *sessions.SessionManager,
	ub *users.UserBackend,
	queue *pubsub.Queue,
	currentRoom *CurrentRoomBackend,
	groups *groups.Backend,
	repository *Repository,
) *Server {
	return &Server{
		sfu:         sfu,
		sm:          sm,
		ub:          ub,
		queue:       queue,
		currentRoom: currentRoom,
		groups:      groups,
		repository:  repository,
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
		// @TODO RETURN?
	}

	user, err := s.userForSession(session)
	if err != nil {
		_ = peer.Close()
		// @TODO RETURN?
	}

	var room *Room

	switch in.Payload.(type) {
	case *pb.SignalRequest_Join:
		join := in.GetJoin()
		if join == nil {
			return status.Errorf(codes.Internal, "something went wrong")
		}

		r, err := s.repository.Get(join.Room)
		if err != nil {
			return status.Errorf(codes.Internal, "join error room closed")
		}

		if r.PeerCount() >= MAX_PEERS {
			return status.Errorf(codes.Internal, "join error room full")
		}

		//if !s.canJoin(user.ID, r) {
		//	return status.Errorf(codes.Internal, "user not invited")
		//}

		description := webrtc.SessionDescription{
			Type: webrtc.NewSDPType(strings.ToLower(join.Description.Type)),
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
					Room: r.ToProtoForPeer(),
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

		session, _ := s.sfu.GetSession(id)
		room = NewRoom(id, name, session) // @TODO THE REST

		description := webrtc.SessionDescription{
			Type: webrtc.NewSDPType(strings.ToLower(create.Description.Type)),
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

		s.repository.Set(room)

	default:
		return status.Error(codes.FailedPrecondition, "invalid message")
	}

	room.Handle(user, peer)

	// @TODO ADD ON DISCONNECT CALLBACK.

	for {
		in, err := receive(peer, stream)
		if err != nil {
			// @TODO close the peer

			if err == io.EOF {
				return nil
			}

			return err
		}

		if in == nil {
			return nil
		}

		err = s.handle(peer, stream, in)
		if err != nil {
			return err
		}
	}
}

func (s *Server) handle(peer *sfu.Peer, stream pb.SFU_SignalServer, in *pb.SignalRequest) error {
	switch in.Payload.(type) {
	case *pb.SignalRequest_Description:
		payload := in.GetDescription()
		sdp := webrtc.SessionDescription{
			Type: webrtc.NewSDPType(strings.ToLower(payload.Type)),
			SDP:  payload.Sdp,
		}

		if sdp.Type == webrtc.SDPTypeOffer {
			answer, err := peer.Answer(sdp)
			if err != nil {
				if err == sfu.ErrNoTransportEstablished || err == sfu.ErrOfferIgnored {
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
				} else {
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
				if err == sfu.ErrNoTransportEstablished {
					err = stream.Send(&pb.SignalReply{
						Payload: &pb.SignalReply_Error{
							Error: fmt.Errorf("set remote description error: %w", err).Error(),
						},
					})

					if err != nil {
						log.Printf("grpc send error %v\n", err)
						return status.Errorf(codes.Internal, err.Error())
					}
				} else {
					return status.Errorf(codes.Unknown, err.Error())
				}
			}
		}
	case *pb.SignalRequest_Trickle:
		payload := in.GetTrickle()

		midLine := uint16(payload.IceCandidate.SdpMLineIndex)
		candidate := webrtc.ICECandidateInit{
			Candidate:     payload.IceCandidate.Candidate,
			SDPMid:        &payload.IceCandidate.SdpMid,
			SDPMLineIndex: &midLine,
		}

		err := peer.Trickle(candidate, int(payload.Target))
		if err != nil {
			if err == sfu.ErrNoTransportEstablished {
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
			} else {
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

	answer, err := peer.Join(room, description)
	if err != nil {
		if err == sfu.ErrTransportExists || err == sfu.ErrOfferIgnored {
			err = stream.Send(&pb.SignalReply{
				Payload: &pb.SignalReply_Error{
					Error: fmt.Errorf("join error: %w", err).Error(),
				},
			})

			if err != nil {
				log.Printf("grpc send error %v ", err)
				return nil, status.Errorf(codes.Internal, err.Error())
			}
		} else {
			return nil, status.Errorf(codes.Unknown, err.Error())
		}
	}

	return answer, nil
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
