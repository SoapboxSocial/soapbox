package rooms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/soapboxsocial/soapbox/pkg/blocks"
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

	sfu     *sfu.SFU
	sm      *sessions.SessionManager
	ub      *users.UserBackend
	groups  *groups.Backend
	queue   *pubsub.Queue
	blocked *blocks.Backend

	currentRoom *CurrentRoomBackend

	rooms map[int]*Room

	nextID int
}

func NewServer(
	sfu *sfu.SFU,
	sm *sessions.SessionManager,
	ub *users.UserBackend,
	queue *pubsub.Queue,
	cr *CurrentRoomBackend,
	groups *groups.Backend,
	blocked *blocks.Backend,
) *Server {
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
		blocked:     blocked,
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

	// @TODO THIS SHOULD BE DONE BETTER.
	blockedUsers, err := s.blocked.GetUsersBlockedBy(id)
	if err != nil {
		fmt.Printf("failed to get blocked users: %+v\n", err)
	}

	rooms := make([]*pb.RoomState, 0)
	for _, r := range s.rooms {
		if !s.canJoin(id, r) {
			continue
		}

		if r.ContainsBlockedUser(blockedUsers) {
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

	isNew := false
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
	case *pb.SignalRequest_Create:
		isNew = true
		s.mux.Lock()
		id := s.nextID

		name := internal.TrimRoomNameToLimit(payload.Create.Name)

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
			payload.Create.Users,
			s.onRoomJoinedEvent,
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

		log.Printf("created room: %d", id)
	default:
		return status.Error(codes.FailedPrecondition, "not joined or created room")
	}

	return room.Handle(user, stream, peer, isNew)
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

func (s *Server) onRoomJoinedEvent(isNew bool, peer int, room *Room) {
	visibility := pubsub.Public
	if room.IsPrivate() {
		visibility = pubsub.Private
	}

	var event pubsub.Event
	group := room.Group()

	if isNew {
		if group == nil {
			event = pubsub.NewRoomCreationEvent(room.Name(), room.id, peer, visibility)
		} else {
			event = pubsub.NewRoomCreationEventWithGroup(room.Name(), room.id, peer, group.ID, visibility)
		}
	} else {
		event = pubsub.NewRoomJoinEvent(room.Name(), room.id, peer, visibility)
	}

	err := s.queue.Publish(pubsub.RoomTopic, event)
	if err != nil {
		log.Printf("queue.Publish err: %v\n", err)
	}

	if visibility == pubsub.Private {
		return
	}

	if group != nil && group.GroupType != "public" {
		return
	}

	err = s.currentRoom.SetCurrentRoomForUser(peer, room.id)
	if err != nil {
		log.Printf("failed to set current room err: %v", err)
	}
}

func (s *Server) getGroup(peer, id int) (*groups.Group, error) {
	isMember, err := s.groups.IsGroupMember(peer, id)
	if err != nil || !isMember {
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
