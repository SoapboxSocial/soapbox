package rooms

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/golang/protobuf/ptypes/empty"
	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/soapboxsocial/soapbox/pkg/notifications"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Server struct {
	mux sync.RWMutex

	sfu   *sfu.SFU
	sm    *sessions.SessionManager
	ub    *users.UserBackend
	queue *notifications.Queue

	currentRoom *CurrentRoomBackend

	rooms map[int]*Room

	nextID int
}

func NewServer(sfu *sfu.SFU, sm *sessions.SessionManager, ub *users.UserBackend, queue *notifications.Queue, cr *CurrentRoomBackend) *Server {
	return &Server{
		mux:         sync.RWMutex{},
		sfu:         sfu,
		sm:          sm,
		ub:          ub,
		queue:       queue,
		currentRoom: cr,
		rooms:       make(map[int]*Room),
		nextID:      1,
	}
}

func (s *Server) ListRooms(context.Context, *empty.Empty) (*pb.RoomList, error) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	rooms := make([]*pb.RoomState, 0)
	for _, r := range s.rooms {
		proto := r.ToProtoForPeer()
		proto.Role = ""
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

	switch payload := in.Payload.(type) {
	case *pb.SignalRequest_Join:
		user, err = s.getMemberForSession(payload.Join.Session)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "unauthenticated")
		}

		s.mux.RLock()
		r, ok := s.rooms[int(payload.Join.Room)]
		s.mux.RUnlock()

		if !ok {
			return status.Errorf(codes.Internal, "join error room closed")
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
		user, err = s.getMemberForSession(payload.Create.Session)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "unauthenticated")
		}

		s.mux.Lock()
		id := s.nextID
		room = NewRoom(id, payload.Create.Name)
		s.nextID++
		s.mux.Unlock()

		room.OnDisconnected(func(room, id int) {
			s.mux.RLock()
			count := s.rooms[room].PeerCount()
			s.mux.RUnlock()

			go func() {
				s.currentRoom.RemoveCurrentRoomForUser(id)
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

		s.queue.Push(notifications.Event{
			Type:    notifications.EventTypeRoomCreation,
			Creator: user.ID,
			Params:  map[string]interface{}{"name": payload.Create.Name, "id": id},
		})

		log.Printf("created room: %d", id)
	default:
		return status.Error(codes.FailedPrecondition, "not joined or created room")
	}

	go func() {
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
	return &member{ID: id, DisplayName: u.DisplayName, Image: u.Image, IsMuted: false, Role: SPEAKER}, nil
}
