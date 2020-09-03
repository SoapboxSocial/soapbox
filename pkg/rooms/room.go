package rooms

import (
	"encoding/json"
	"io"
	"log"
	"sync"

	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type PeerRole string

const (
	OWNER    PeerRole = "owner"
	SPEAKER  PeerRole = "speaker"
	AUDIENCE PeerRole = "audience"
)

// member is used to communicate what peers are part of the chat
type member struct {
	ID          int      `json:"id"`
	DisplayName string   `json:"display_name"`
	Image       string   `json:"image"`
	Role        PeerRole `json:"role"`
	IsMuted     bool     `json:"is_muted"`
}

type payload struct {
	ID      int      `json:"id"`
	Name    string   `json:"name,omitempty"`
	Members []member `json:"members"`
}

type peer struct {
	me     *member
	stream pb.RoomService_SignalServer
	rtc    *sfu.WebRTCTransport
}

type Room struct {
	mux sync.RWMutex

	id   int
	name string

	members map[int]*peer
}

func NewRoom(id int, name string) *Room {
	return &Room{
		mux:     sync.RWMutex{},
		id:      id,
		name:    name,
		members: make(map[int]*peer),
	}
}

func (r *Room) Handle(id int, stream pb.RoomService_SignalServer, rtc *sfu.WebRTCTransport) error {
	r.mux.Lock()
	r.members[id] = &peer{
		stream: stream,
		rtc:    rtc,
	}
	r.mux.Unlock()

	me := member{ID: id, DisplayName: "foo", Image: "", Role: SPEAKER, IsMuted: false}
	data, err := json.Marshal(me)
	if err != nil {
		log.Printf("failed to encode: %s\n", err.Error())
	}

	r.notify(&pb.SignalReply_Event{
		Type: pb.SignalReply_Event_JOINED,
		From: int64(id),
		Data: data,
	})

	for {
		in, err := stream.Recv()
		if err != nil {
			// @TODO: Potentially change owner
			// @TODO: Close room if last disconnect
			go r.notify(&pb.SignalReply_Event{
				Type: pb.SignalReply_Event_LEFT,
				From: int64(id),
			})

			_ = rtc.Close()

			r.mux.Lock()
			delete(r.members, id)
			r.mux.Unlock()

			if err == io.EOF {
				return nil
			}

			errStatus, _ := status.FromError(err)
			if errStatus.Code() == codes.Canceled {
				return nil
			}

			log.Printf("signal error %v %v\n", errStatus.Message(), errStatus.Code())
			return err
		}

		err = r.onPayload(id, in)
		if err != nil {
			return err
		}
	}
}

func (r *Room) onPayload(from int, in *pb.SignalRequest) error {
	switch payload := in.Payload.(type) {
	case *pb.SignalRequest_Negotiate:
		if payload.Negotiate.Type != "answer" {
			// @todo
			return nil
		}

		r.mux.RLock()
		peer := r.members[from]
		r.mux.RUnlock()

		err := peer.rtc.SetRemoteDescription(webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  string(payload.Negotiate.Sdp),
		})

		if err != nil {
			return status.Errorf(codes.Internal, "%s", err)
		}
	case *pb.SignalRequest_Trickle:
		var candidate webrtc.ICECandidateInit
		err := json.Unmarshal([]byte(payload.Trickle.Init), &candidate)
		if err != nil {
			log.Printf("error parsing ice candidate: %v", err)
		}

		r.mux.RLock()
		peer := r.members[from]
		r.mux.RUnlock()

		err = peer.rtc.AddICECandidate(candidate)
		if err != nil {
			return status.Errorf(codes.Internal, "error adding ice candidate")
		}
	case *pb.SignalRequest_Command_:
		return r.onCommand(from, payload.Command)
	}

	return nil
}

func (r *Room) onCommand(from int, cmd *pb.SignalRequest_Command) error {
	switch cmd.Type {
	case pb.SignalRequest_Command_ADD_SPEAKER:
		// @TODO IN DEVELOPMENT
		break
	case pb.SignalRequest_Command_REMOVE_SPEAKER:
		// @TODO IN DEVELOPMENT
		break
	case pb.SignalRequest_Command_MUTE_SPEAKER:
		r.mux.Lock()
		// @TODO SET MUTED
		r.mux.Unlock()

		go r.notify(&pb.SignalReply_Event{
			Type: pb.SignalReply_Event_MUTED_SPEAKER,
			From: int64(from),
		})
	case pb.SignalRequest_Command_UNMUTE_SPEAKER:
		r.mux.Lock()
		// @TODO SET UNMUTED
		r.mux.Unlock()

		go r.notify(&pb.SignalReply_Event{
			Type: pb.SignalReply_Event_UNMUTED_SPEAKER,
			From: int64(from),
		})
	case pb.SignalRequest_Command_REACTION:
		go r.notify(&pb.SignalReply_Event{
			Type: pb.SignalReply_Event_REACTED,
			From: int64(from),
			Data: cmd.Data,
		})
	}

	return nil
}

func (r *Room) notify(event *pb.SignalReply_Event) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	for id, p := range r.members {
		if int64(id) == event.From {
			continue
		}

		err := p.stream.Send(&pb.SignalReply{
			Payload: &pb.SignalReply_Event_{
				Event: event,
			},
		})

		if err != nil {
			// @todo
			log.Printf("failed to write to data channel: %s\n", err.Error())
		}
	}
}

func (r *Room) MarshalJSON() ([]byte, error) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	payload := payload{
		ID:      r.id,
		Name:    r.name,
		Members: make([]member, 0),
	}

	for _, p := range r.members {
		payload.Members = append(payload.Members, *p.me)
	}

	return json.Marshal(payload)
}

func (r *Room) ToProtoForPeer() *pb.RoomState {
	r.mux.RLock()
	defer r.mux.RUnlock()

	members := make([]*pb.RoomState_RoomMember, 0)

	for _, member := range r.members {
		members = append(members, &pb.RoomState_RoomMember{
			Id:          int64(member.me.ID),
			DisplayName: member.me.DisplayName,
			Image:       member.me.Image,
			Role:        string(member.me.Role),
			Muted:       member.me.IsMuted,
		})
	}

	return &pb.RoomState{
		Name:    r.name,
		Role:    string(SPEAKER), // @TODO THIS SHOULD DEPEND ON WHO OWNS THE ROOM ETC
		Members: members,
	}
}
