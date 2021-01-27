package rooms

import (
	"fmt"
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

	session, err := internal.SessionID(r)
	if err != nil {
		_ = conn.Close()
		return
	}

	user, err := s.userForSession(session)
	if err != nil {
		_ = conn.Close()
		return
	}

	peer := sfu.NewPeer(s.sfu)
	me := NewMember(user.ID, user.DisplayName, user.Image, peer, conn)

	in, err := me.ReceiveMsg()
	if err != nil {
		log.Printf("receive err: %v", err)
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

		answer, err := peer.Join(join.Room, description)
		if err != nil && (err != sfu.ErrTransportExists && err != sfu.ErrOfferIgnored) {
			log.Printf("join err: %v", err)
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
		//if create.GetGroup() != 0 {
		//	group, err = s.getGroup(user.ID, int(create.GetGroup()))
		//	if err != nil {
		//		fmt.Printf("group err: %v", err)
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

		answer, err := peer.Join(id, description)
		if err != nil && (err != sfu.ErrTransportExists && err != sfu.ErrOfferIgnored) {
			log.Printf("create err: %v", err)
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

	// @TODO ONCE WE SWITCH THE TRANSPORT WE WILL BE ABLE TO RELEASE HERE
	room.Handle(me)
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
