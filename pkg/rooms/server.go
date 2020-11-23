package rooms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/soapboxsocial/soapbox/pkg/groups"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
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

	rooms map[int]*Room

	nextID int
}

func NewServer(sfu *sfu.SFU, sm *sessions.SessionManager, ub *users.UserBackend, queue *pubsub.Queue, cr *CurrentRoomBackend, groups *groups.Backend) *Server {
	return &Server{
		mux:         sync.RWMutex{},
		sfu:         sfu,
		sm:          sm,
		ub:          ub,
		queue:       queue,
		currentRoom: cr,
		rooms:       make(map[int]*Room),
		nextID:      1,
		groups:      groups,
	}
}

// Deprecated: Remove
func (s *Server) ListRooms(ctx context.Context, _ *empty.Empty) (*pb.RoomList, error) {
	auth, err := authForContext(ctx)
	if err != nil {
		return nil, err
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	id, err := s.sm.GetUserIDForSession(auth)
	if err != nil {
		return nil, err
	}

	rooms := make([]*pb.RoomState, 0)
	for _, r := range s.rooms {
		if !s.canJoin(id, r) {
			continue
		}

		proto := r.ToProtoForPeer()
		proto.Role = ""

		if len(proto.Members) == 0 {
			continue
		}

		rooms = append(rooms, proto)
	}

	return &pb.RoomList{Rooms: rooms}, nil
}

func (s *Server) Signal(stream pb.RoomService_SignalServer) error {
	in, err := stream.Recv()
	if err != nil {
		return err
	}

	var room *Room
	var peer *sfu.WebRTCTransport
	var user *member

	auth, _ := authForContext(stream.Context())
	user, err = s.getMemberForSession(auth)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "unauthenticated")
	}

	// Auth will be able to be moved out.

	switch payload := in.Payload.(type) {
	case *pb.SignalRequest_Join:
		s.mux.RLock()
		r, ok := s.rooms[int(payload.Join.Room)]
		s.mux.RUnlock()

		if !ok {
			return status.Errorf(codes.Internal, "join error room closed")
		}

		if r.PeerCount() >= MAX_PEERS {
			return status.Errorf(codes.Internal, "join error room full")
		}

		if !s.canJoin(user.ID, r) {
			return status.Errorf(codes.Internal, "user not invited")
		}

		room = r
		proto := r.ToProtoForPeer()

		peer, err = s.setupConnection(int(payload.Join.Room), stream)
		if err != nil {
			return status.Errorf(codes.Internal, "join error %s", err)
		}

		offer := peer.LocalDescription()
		err = stream.Send(&pb.SignalReply{
			Payload: &pb.SignalReply_Join{
				Join: &pb.JoinReply{
					Room: proto,
					Answer: &pb.SessionDescription{
						Type: offer.Type.String(),
						Sdp:  []byte(offer.SDP),
					},
				},
			},
		})

		if err != nil {
			log.Printf("error sending join response %s", err)
			return status.Errorf(codes.Internal, "join error %s", err)
		}

		visibility := pubsub.Public
		if r.IsPrivate() {
			visibility = pubsub.Private
		}

		err := s.queue.Publish(
			pubsub.RoomTopic,
			pubsub.NewRoomJoinEvent(
				r.Name(),
				int(payload.Join.Room),
				user.ID,
				visibility,
			),
		)

		if err != nil {
			log.Printf("queue.Publish err: %v\n", err)
		}
	case *pb.SignalRequest_Create:
		s.mux.Lock()
		id := s.nextID

		name := strings.TrimSpace(payload.Create.Name)
		if len([]rune(name)) > 30 {
			name = string([]rune(name)[:30])
		}

		var group *groups.Group
		if payload.Create.GetGroup() != 0 {
			group, err = s.getGroup(user.ID, int(payload.Create.GetGroup()))
			if err != nil {
				return status.Errorf(codes.Internal, "group error %s", err)
			}
		}

		room = NewRoom(
			id,
			name,
			s.queue,
			payload.Create.GetVisibility() == pb.Visibility_PRIVATE,
			user.ID,
			group,
		)
		s.nextID++
		s.mux.Unlock()

		room.OnDisconnected(func(room, id int) {
			s.mux.RLock()
			r := s.rooms[room]
			s.mux.RUnlock()

			if r == nil {
				return
			}

			count := r.PeerCount()

			go func() {
				s.currentRoom.RemoveCurrentRoomForUser(id)

				err := s.queue.Publish(pubsub.RoomTopic, pubsub.NewRoomLeftEvent(room, id))
				if err != nil {
					log.Printf("queue.Publish err: %v\n", err)
				}
			}()

			if count > 0 {
				return
			}

			s.mux.Lock()
			delete(s.rooms, room)
			s.mux.Unlock()

			log.Printf("room %d was closed", room)
		})

		peer, err = s.setupConnection(id, stream)
		if err != nil {
			return status.Errorf(codes.Internal, "join error %s", err)
		}

		offer := peer.LocalDescription()
		err = stream.Send(&pb.SignalReply{
			Payload: &pb.SignalReply_Create{
				Create: &pb.CreateReply{
					Id: int64(id),
					Answer: &pb.SessionDescription{
						Type: offer.Type.String(),
						Sdp:  []byte(offer.SDP),
					},
				},
			},
		})

		if err != nil {
			log.Printf("error sending join response %s", err)
			return status.Errorf(codes.Internal, "join error %s", err)
		}

		// We only add the room when its safely created
		s.mux.Lock()
		s.rooms[id] = room
		s.mux.Unlock()

		// @TODO CHECK GROUP
		visibility := pubsub.Public
		if payload.Create.Visibility == pb.Visibility_PRIVATE {
			visibility = pubsub.Private
		}

		err := s.queue.Publish(
			pubsub.RoomTopic,
			pubsub.NewRoomCreationEvent(
				payload.Create.Name,
				id,
				user.ID,
				visibility,
			),
		)

		if err != nil {
			log.Printf("queue.Publish err: %v\n", err)
		}

		log.Printf("created room: %d", id)
	default:
		return status.Error(codes.FailedPrecondition, "not joined or created room")
	}

	go func() {
		// @TODO THIS IS TEMP
		// @TODO WE NEED TO CHECK GROUP HERE
		if room.IsPrivate() {
			return
		}

		err := s.currentRoom.SetCurrentRoomForUser(user.ID, room.id)
		if err != nil {
			log.Printf("failed to set current room err: %v", err)
		}
	}()

	return room.Handle(user, stream, peer)
}

func (s *Server) setupConnection(room int, stream pb.RoomService_SignalServer) (*sfu.WebRTCTransport, error) {
	me := sfu.MediaEngine{}
	me.RegisterDefaultCodecs()

	peer, err := s.sfu.NewWebRTCTransport(string(room), me)
	if err != nil {
		return nil, err
	}

	_, err = peer.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio)
	if err != nil {
		return nil, err
	}

	offer, err := peer.CreateOffer()
	if err != nil {
		return nil, err
	}

	err = peer.SetLocalDescription(offer)
	if err != nil {
		return nil, err
	}

	// Notify user of trickle candidates
	peer.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		data, err := json.Marshal(c.ToJSON())
		if err != nil {
			log.Printf("json marshal candidate error: %v\n", err)
			return
		}

		err = stream.Send(&pb.SignalReply{
			Payload: &pb.SignalReply_Trickle{
				Trickle: &pb.Trickle{
					Init: string(data),
				},
			},
		})

		if err != nil {
			log.Printf("OnIceCandidate error %s", err)
		}
	})

	peer.OnNegotiationNeeded(func() {
		log.Println("on negotiation needed called")
		offer, err := peer.CreateOffer()
		if err != nil {
			log.Printf("CreateOffer error: %v\n", err)
			return
		}

		err = peer.SetLocalDescription(offer)
		if err != nil {
			log.Printf("SetLocalDescription error: %v\n", err)
			return
		}

		err = stream.Send(&pb.SignalReply{
			Payload: &pb.SignalReply_Negotiate{
				Negotiate: &pb.SessionDescription{
					Type: offer.Type.String(),
					Sdp:  []byte(offer.SDP),
				},
			},
		})

		if err != nil {
			log.Printf("negotiation error %s", err)
		}
	})

	return peer, nil
}

func (s *Server) getGroup(peer, id int) (*groups.Group, error) {
	isMember, err := s.groups.IsGroupMember(peer, id)
	if err != nil || isMember == false {
		return nil, fmt.Errorf("user %d is not member of group %d", peer, id)
	}

	return s.groups.FindById(id)
}

func (s *Server) canJoin(peer int, room *Room) bool {
	group := room.Group()
	if group == nil {
		return room.CanJoin(peer)
	}

	if group.GroupType == "public" {
		return true
	}

	isMember, err := s.groups.IsGroupMember(peer, group.ID)
	if err != nil {
		return false
	}

	return isMember
}

func (s *Server) getMemberForSession(session string) (*member, error) {
	id, err := s.sm.GetUserIDForSession(session)
	if err != nil {
		return nil, err
	}

	u, err := s.ub.FindByID(id)
	if err != nil {
		return nil, err
	}

	// @TODO ROLE SHOULD BE BASED ON STUFF
	return &member{ID: id, DisplayName: u.DisplayName, Image: u.Image, IsMuted: true, Role: SPEAKER}, nil
}

func authForContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("missing metadata")
	}

	auth := md.Get("authorization")
	if len(auth) != 1 {
		return "", errors.New("unauthorized")
	}

	return auth[0], nil
}
