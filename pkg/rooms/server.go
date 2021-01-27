package rooms

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"

	"github.com/soapboxsocial/soapbox/pkg/groups"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms/internal"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/rooms/signal"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

const MAX_PEERS = 16

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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

func (s *Server) Signal(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	conn := signal.NewWebSocketTransport(c)

	peer := sfu.NewPeer(s.sfu)

	in, err := receive(peer, conn)
	if err != nil {
		log.Printf("receive err: %v", err)
		_ = conn.Close()
		_ = peer.Close()
		return
	}

	if in == nil {
		_ = conn.Close()
		_ = peer.Close()
		return
	}

	session, err := internal.SessionID(r)
	if err != nil {
		_ = conn.Close()
		_ = peer.Close()
		return
	}

	user, err := s.userForSession(session)
	if err != nil {
		_ = conn.Close()
		_ = peer.Close()
		return
	}

	var room *Room

	switch in.Payload.(type) {
	case *pb.SignalRequest_Join:
		join := in.GetJoin()
		if join == nil {
			return
		}

		r, err := s.repository.Get(join.Room)
		if err != nil {
			return
		}

		if r.PeerCount() >= MAX_PEERS {
			return
		}

		//if !s.canJoin(user.ID, r) {
		//	return status.Errorf(codes.Internal, "user not invited")
		//}

		description := webrtc.SessionDescription{
			Type: webrtc.NewSDPType(strings.ToLower(join.Description.Type)),
			SDP:  join.Description.Sdp,
		}

		answer, err := setup(peer, join.Room, conn, description)
		if err != nil {
			log.Printf("setup err: %v", err)
			return
		}

		err = conn.Write(&pb.SignalReply{
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
			return
		}

		room = r
	case *pb.SignalRequest_Create:
		create := in.GetCreate()
		if create == nil {
			return
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
		room = NewRoom(id, name, create.Visibility, session) // @TODO THE REST, ENSURE USER IS MARKED ADMIN

		room.OnDisconnected(func(room string, id int) {
			r, err := s.repository.Get(room)
			if err != nil {
				fmt.Printf("failed to get room %v\n", err)
				return
			}

			go func() {
				s.currentRoom.RemoveCurrentRoomForUser(id)

				err := s.queue.Publish(pubsub.RoomTopic, pubsub.NewRoomLeftEvent(room, id))
				if err != nil {
					log.Printf("queue.Publish err: %v\n", err)
				}
			}()

			if r.PeerCount() > 0 {
				return
			}

			s.repository.Remove(room)

			log.Printf("room \"%s\" was closed", room)
		})

		description := webrtc.SessionDescription{
			Type: webrtc.NewSDPType(strings.ToLower(create.Description.Type)),
			SDP:  create.Description.Sdp,
		}

		answer, err := setup(peer, id, conn, description)
		if err != nil {
			return
		}

		err = conn.Write(&pb.SignalReply{
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
			return
		}

		// @TODO ensure room isn't shown when peer is not yet connected.
		s.repository.Set(room)

	default:
		return
	}

	room.Handle(user, peer)

	// @TODO: HANDLE IF SIGNAL CLIENT DISCONNECTS

	for {
		in, err := receive(peer, conn)
		if err != nil {
			return
		}

		if in == nil {
			return
		}

		err = s.handle(peer, conn, in)
		if err != nil {
			return
		}
	}
}

func (s *Server) handle(peer *sfu.Peer, conn signal.Transport, in *pb.SignalRequest) error {
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
					return nil
				}

				return fmt.Errorf("negotatie err: %v", err)
			}

			err = conn.Write(&pb.SignalReply{
				Id: in.Id,
				Payload: &pb.SignalReply_Description{
					Description: &pb.SessionDescription{
						Type: answer.Type.String(),
						Sdp:  answer.SDP,
					},
				},
			})

			if err != nil {
				log.Printf("conn.Write failed: %v", err)
				return err
			}

		} else if sdp.Type == webrtc.SDPTypeAnswer {
			err := peer.SetRemoteDescription(sdp)
			if err != nil && err != sfu.ErrNoTransportEstablished {
				return err
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
		if err != nil && err != sfu.ErrNoTransportEstablished {
			return fmt.Errorf("negotatie err: %v", err)
		}
	}

	return nil
}

func receive(peer *sfu.Peer, conn signal.Transport) (*pb.SignalRequest, error) {
	msg, err := conn.ReadMsg()
	if err != nil {
		_ = peer.Close()

		if err == io.EOF {
			return nil, nil
		}

		log.Printf("signal error %v", err)
		return nil, err
	}

	return msg, nil
}

// setup properly sets up a peer to communicate with the SFU.
// @TODO, here we should ideally pass the room member. this has a signalling transport property.
// when webrtc is fully connected it will switch signalling from GRPC to webrtc.
func setup(peer *sfu.Peer, room string, conn signal.Transport, description webrtc.SessionDescription) (*webrtc.SessionDescription, error) {

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

		err := conn.Write(&pb.SignalReply{
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

		err := conn.Write(&pb.SignalReply{
			Payload: &pb.SignalReply_Description{
				Description: description,
			},
		})

		if err != nil {
			log.Printf("negotiation error %s", err)
		}
	}

	answer, err := peer.Join(room, description)
	if err != nil && (err != sfu.ErrTransportExists && err != sfu.ErrOfferIgnored) {
		return nil, err
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
