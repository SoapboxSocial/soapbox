package rooms

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"

	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
)

type Server struct {
	mux sync.RWMutex

	sfu *sfu.SFU
	sm  *sessions.SessionManager

	rooms map[int]*Room

	nextID int
}

func NewServer(sfu *sfu.SFU) *Server {
	s := &Server{
		mux:    sync.RWMutex{},
		sfu:    sfu,
		rooms:  make(map[int]*Room),
		nextID: 2,
	}

	s.rooms[1] = NewRoom(1, "foo")
	return s
}

func (s *Server) Signal(stream pb.RoomService_SignalServer) error {
	in, err := stream.Recv()
	if err != nil {
		return err
	}

	var room *Room
	var peer *sfu.WebRTCTransport

	// @todo check session and shit

	// @todo get random id to tests
	id := rand.Int()

	switch payload := in.Payload.(type) {
	case *pb.SignalRequest_Join:
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
		s.mux.Lock()
		id := s.nextID
		room = NewRoom(id, payload.Create.Name)
		s.rooms[id] = room
		s.nextID++
		s.mux.Unlock()

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
	default:
		return status.Error(codes.FailedPrecondition, "not joined or created room")
	}

	return room.Handle(id, stream, peer)
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
