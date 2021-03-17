package rooms

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"

	"github.com/soapboxsocial/soapbox/pkg/blocks"
	"github.com/soapboxsocial/soapbox/pkg/groups"
	httputil "github.com/soapboxsocial/soapbox/pkg/http"
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
	sfu     *sfu.SFU
	sm      *sessions.SessionManager
	ub      *users.UserBackend
	groups  *groups.Backend
	queue   *pubsub.Queue
	blocked *blocks.Backend

	ws          *WelcomeStore
	currentRoom *CurrentRoomBackend

	repository *Repository
}

func NewServer(
	sfu *sfu.SFU,
	sm *sessions.SessionManager,
	ub *users.UserBackend,
	queue *pubsub.Queue,
	currentRoom *CurrentRoomBackend,
	ws *WelcomeStore,
	groups *groups.Backend,
	repository *Repository,
	blocked *blocks.Backend,
) *Server {
	return &Server{
		sfu:         sfu,
		sm:          sm,
		ub:          ub,
		queue:       queue,
		currentRoom: currentRoom,
		ws:          ws,
		groups:      groups,
		repository:  repository,
		blocked:     blocked,
	}
}

func (s *Server) Signal(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	conn := signal.NewWebSocketTransport(c)

	user, err := s.userForSession(r)
	if err != nil {
		_ = conn.Close()
		return
	}

	peer := sfu.NewPeer(s.sfu)
	me := NewMember(user.ID, user.DisplayName, user.Image, peer, conn)

	in, err := me.ReceiveMsg()
	if err != nil {
		log.Printf("receive err: %v", err)
		_ = conn.Close()
		return
	}

	var room *Room

	switch in.Payload.(type) {
	case *pb.SignalRequest_Join:
		join := in.GetJoin()
		if join == nil {
			return
		}

		// @TODO COULD PROBABLY CLEAN THIS UP WITH A PRECONDITION, if !repo.has -> check if welcome room -> add

		r, err := s.getRoom(join.Room, user.ID)
		if err != nil {
			_ = conn.WriteError(in.Id, pb.SignalReply_ERROR_CLOSED)
			return
		}

		if r.PeerCount() >= MAX_PEERS {
			_ = conn.WriteError(in.Id, pb.SignalReply_ERROR_FULL)
			return
		}

		if !s.canJoin(user.ID, r) {
			_ = conn.WriteError(in.Id, pb.SignalReply_ERROR_NOT_INVITED)
			return
		}

		err = peer.Join(join.Room, strconv.Itoa(user.ID))
		if err != nil && (err != sfu.ErrTransportExists && err != sfu.ErrOfferIgnored) {
			_ = conn.WriteError(in.Id, pb.SignalReply_ERROR_CLOSED)
			return
		}

		description := webrtc.SessionDescription{
			Type: webrtc.NewSDPType(strings.ToLower(join.Description.Type)),
			SDP:  join.Description.Sdp,
		}

		answer, err := peer.Answer(description)
		if err != nil {
			_ = conn.WriteError(in.Id, pb.SignalReply_ERROR_CLOSED)
			return
		}

		if answer == nil {
			_ = conn.WriteError(in.Id, pb.SignalReply_ERROR_CLOSED)
			return
		}

		if r.WasAdminOnDisconnect(user.ID) {
			me.SetRole(pb.RoomState_RoomMember_ROLE_ADMIN)
		}

		err = conn.Write(&pb.SignalReply{
			Id: in.Id,
			Payload: &pb.SignalReply_Join{
				Join: &pb.JoinReply{
					Room: r.ToProto(),
					Description: &pb.SessionDescription{
						Type: answer.Type.String(),
						Sdp:  answer.SDP,
					},
					Role: me.Role(),
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

		var group *groups.Group
		if create.Visibility != pb.Visibility_VISIBILITY_PRIVATE && create.GetGroup() != 0 {
			group, err = s.getGroup(user.ID, int(create.GetGroup()))
			if err != nil {
				fmt.Printf("group err: %v", err)
			}
		}

		room = s.createRoom(
			id,
			name,
			me.id,
			create.Visibility,
			group,
		)

		// @TODO SHOULD PROBABLY BE IN A CALLBACK SO WE KNOW THE ROOM IS OPEN
		if create.Visibility == pb.Visibility_VISIBILITY_PRIVATE {
			for _, id := range create.Users {
				room.InviteUser(me.id, int(id))
			}
		}

		err = peer.Join(id, strconv.Itoa(user.ID))
		if err != nil && (err != sfu.ErrTransportExists && err != sfu.ErrOfferIgnored) {
			_ = conn.WriteError(in.Id, pb.SignalReply_ERROR_CLOSED)
			return
		}

		description := webrtc.SessionDescription{
			Type: webrtc.NewSDPType(strings.ToLower(create.Description.Type)),
			SDP:  create.Description.Sdp,
		}

		answer, err := peer.Answer(description)
		if err != nil {
			_ = conn.WriteError(in.Id, pb.SignalReply_ERROR_CLOSED)
			return
		}

		if answer == nil {
			_ = conn.WriteError(in.Id, pb.SignalReply_ERROR_CLOSED)
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

		s.repository.Set(room)

	default:
		return
	}

	// @TODO ONCE WE SWITCH THE TRANSPORT WE WILL BE ABLE TO RELEASE HERE
	room.Handle(me)
}

func (s *Server) getRoom(id string, owner int) (*Room, error) {
	r, err := s.repository.Get(id)
	if err == nil {
		return r, nil
	}

	user, err := s.ws.GetUserIDForWelcomeRoom(id)
	if err != nil {
		return nil, err
	}

	if user == 0 {
		return nil, errors.New("unknown room")
	}

	// @TODO NAME
	r = s.createRoom(id, "Welcome!", owner, pb.Visibility_VISIBILITY_PUBLIC, nil)
	s.repository.Set(r)

	r.InviteUser(owner, user)

	return r, nil
}

func (s *Server) createRoom(id, name string, owner int, visibility pb.Visibility, group *groups.Group) *Room {
	session, _ := s.sfu.GetSession(id)

	room := NewRoom(id, name, group, owner, visibility, session, s.queue)

	room.OnDisconnected(func(room string, id int) {
		go func() {
			err := s.currentRoom.RemoveCurrentRoomForUser(id)
			if err != nil {
				log.Printf("currentRoom.RemoveCurrentRoomForUser err: %v\n", err)
			}

			err = s.queue.Publish(pubsub.RoomTopic, pubsub.NewRoomLeftEvent(room, id))
			if err != nil {
				log.Printf("queue.Publish err: %v\n", err)
			}
		}()

		r, err := s.repository.Get(room)
		if err != nil {
			fmt.Printf("failed to get room %v\n", err)
			return
		}

		if r.PeerCount() > 0 {
			return
		}

		s.repository.Remove(room)

		log.Printf("room \"%s\" was closed", room)
	})

	room.OnInvite(func(room string, from, to int) {
		r, err := s.repository.Get(room)
		if err != nil {
			return
		}

		err = s.queue.Publish(pubsub.RoomTopic, pubsub.NewRoomInviteEvent(r.Name(), r.id, from, to))
		if err != nil {
			log.Printf("queue.Publish err: %v\n", err)
		}
	})

	room.OnJoin(func(room *Room, me *Member, isNew bool) {
		visibility := pubsub.Public
		if room.Visibility() == pb.Visibility_VISIBILITY_PRIVATE {
			visibility = pubsub.Private
		}

		var event pubsub.Event
		group := room.Group()

		if isNew {
			if group == nil {
				event = pubsub.NewRoomCreationEvent(room.id, me.id, visibility)
			} else {
				event = pubsub.NewRoomCreationEventWithGroup(room.id, me.id, group.ID, visibility)
			}
		} else {
			event = pubsub.NewRoomJoinEvent(room.id, me.id, visibility)
		}

		err := s.queue.Publish(pubsub.RoomTopic, event)
		if err != nil {
			log.Printf("queue.Publish err: %v\n", err)
		}

		if visibility == pubsub.Private {
			return
		}

		if group != nil && group.GroupType == "private" {
			return
		}

		err = s.currentRoom.SetCurrentRoomForUser(me.id, room.id)
		if err != nil {
			log.Printf("failed to set current room err: %v", err)
		}
	})

	return room
}

func (s *Server) canJoin(peer int, room *Room) bool {
	if !room.CanJoin(peer) {
		return false
	}

	blockingUsers, err := s.blocked.GetUsersWhoBlocked(peer)
	if err != nil {
		fmt.Printf("failed to get blocked users who blocked: %+v", err)
	}

	if room.ContainsUsers(blockingUsers) {
		return false
	}

	group := room.Group()
	if group == nil {
		return true
	}

	if group.GroupType != "private" {
		return true
	}

	isMember, err := s.groups.IsGroupMember(peer, group.ID)
	if err != nil {
		return false
	}

	return isMember
}

func (s *Server) getGroup(peer, id int) (*groups.Group, error) {
	isMember, err := s.groups.IsGroupMember(peer, id)
	if err != nil || !isMember {
		return nil, fmt.Errorf("user %d is not member of group %d", peer, id)
	}

	return s.groups.FindById(id)
}

func (s *Server) userForSession(r *http.Request) (*users.User, error) {
	userID, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		return nil, errors.New("not authenticated")
	}

	u, err := s.ub.FindByID(userID)
	if err != nil {
		return nil, err
	}

	return u, nil
}
