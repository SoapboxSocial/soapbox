package rooms

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strings"
	"sync"

	sfu "github.com/pion/ion-sfu/pkg"
	"github.com/pion/webrtc/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/soapboxsocial/soapbox/pkg/groups"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type PeerRole string

const (
	ADMIN    PeerRole = "admin"
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
	SSRC        uint32   `json:"ssrc"`
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

type OnJoinHandler func(isNew bool, peer int, room *Room)

type Room struct {
	mux sync.RWMutex

	id    int
	name  string
	group *groups.Group

	members map[int]*peer

	onDisconnectedHandlerFunc func(room, peer int)

	queue *pubsub.Queue

	isPrivate bool
	invited   map[int]bool
	kicked    map[int]bool

	onJoin OnJoinHandler
}

func NewRoom(id int, name string, queue *pubsub.Queue, isPrivate bool, owner int, group *groups.Group, onJoin OnJoinHandler) *Room {
	r := &Room{
		mux:       sync.RWMutex{},
		id:        id,
		name:      name,
		members:   make(map[int]*peer),
		queue:     queue,
		isPrivate: isPrivate,
		invited:   make(map[int]bool),
		kicked:    make(map[int]bool),
		group:     group,
		onJoin:    onJoin,
	}

	r.invited[owner] = true

	return r
}

func (r *Room) ID() int {
	return r.id
}

func (r *Room) Group() *groups.Group {
	return r.group
}

func (r *Room) IsPrivate() bool {
	return r.isPrivate
}

func (r *Room) CanJoin(id int) bool {
	r.mux.RLock()
	defer r.mux.RUnlock()

	if r.isPrivate {
		return r.invited[id]
	}

	return !r.kicked[id]
}

func (r *Room) Name() string {
	r.mux.RLock()
	defer r.mux.RUnlock()

	return r.name
}

func (r *Room) PeerCount() int {
	r.mux.RLock()
	defer r.mux.RUnlock()

	return len(r.members)
}

func (r *Room) OnDisconnected(f func(room, peer int)) {
	r.onDisconnectedHandlerFunc = f
}

func (r *Room) Handle(me *member, stream pb.RoomService_SignalServer, rtc *sfu.WebRTCTransport, isNew bool) error {
	id := me.ID

	log.Printf("peer %d joined %d", id, r.id)

	r.mux.Lock()
	_, ok := r.members[id]
	if ok {
		r.mux.Unlock()
		return errors.New("user tried to double enter")
	}
	r.mux.Unlock()

	if r.PeerCount() == 0 {
		me.Role = ADMIN
	}

	r.mux.Lock()
	r.members[id] = &peer{
		me:     me,
		stream: stream,
		rtc:    rtc,
	}
	r.mux.Unlock()

	done := make(chan bool)
	rtc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		if s == webrtc.PeerConnectionStateClosed || s == webrtc.PeerConnectionStateFailed {
			r.onDisconnected(id)
			done <- true
		}
	})

	rtc.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver) {
		r.mux.Lock()
		r.members[id].me.SSRC = track.SSRC()
		r.mux.Unlock()

		r.mux.RLock()
		me := r.members[id]
		r.mux.RUnlock()

		data, err := json.Marshal(me.me)
		if err != nil {
			log.Printf("failed to encode: %s\n", err.Error())
		}

		r.notify(&pb.SignalReply_Event{
			Type: pb.SignalReply_Event_JOINED,
			From: int64(id),
			Data: data,
		})

		go r.onJoin(isNew, id, r)
	})

	for {
		select {
		case <-done:
			return nil
		default:
			in, err := stream.Recv()
			if err != nil {
				r.onDisconnected(id)

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
}

func (r *Room) onDisconnected(peer int) {
	r.mux.RLock()
	p, ok := r.members[peer]
	r.mux.RUnlock()

	if !ok {
		return
	}

	go r.notify(&pb.SignalReply_Event{
		Type: pb.SignalReply_Event_LEFT,
		From: int64(peer),
	})

	err := p.rtc.Close()
	if err != nil {
		log.Printf("rtc.Close error %v\n", err)
	}

	r.mux.Lock()
	delete(r.members, peer)
	r.mux.Unlock()

	r.electRandomAdmin(peer)

	r.onDisconnectedHandlerFunc(r.id, peer)
}

func (r *Room) electRandomAdmin(previous int) {
	r.mux.Lock()
	defer r.mux.Unlock()

	hasAdmin := has(r.members, func(p *peer) bool {
		return p.me.Role == ADMIN
	})

	if hasAdmin {
		return
	}

	for k := range r.members {
		r.members[k].me.Role = ADMIN

		go r.notify(&pb.SignalReply_Event{
			Type: pb.SignalReply_Event_ADDED_ADMIN,
			From: int64(previous),
			Data: intToBytes(k),
		})
		break
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
		peer, ok := r.members[from]
		r.mux.RUnlock()

		if !ok {
			return status.Errorf(codes.Internal, "peer not found")
		}

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
		peer, ok := r.members[from]
		r.mux.RUnlock()
		if !ok {
			return status.Errorf(codes.Internal, "peer not found")
		}

		err = peer.rtc.AddICECandidate(candidate)
		if err != nil {
			return status.Errorf(codes.Internal, "error adding ice candidate")
		}
	case *pb.SignalRequest_Command_:
		return r.onCommand(from, payload.Command)
	case *pb.SignalRequest_Invite:
		return r.onInvite(from, payload.Invite)
	case *pb.SignalRequest_Kick:
		return r.onKick(from, payload.Kick)
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
		_, ok := r.members[from]
		if ok {
			r.members[from].me.IsMuted = true
		}
		r.mux.Unlock()

		go r.notify(&pb.SignalReply_Event{
			Type: pb.SignalReply_Event_MUTED_SPEAKER,
			From: int64(from),
		})
	case pb.SignalRequest_Command_UNMUTE_SPEAKER:
		r.mux.Lock()
		_, ok := r.members[from]
		if ok {
			r.members[from].me.IsMuted = false
		}
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
	case pb.SignalRequest_Command_LINK_SHARE:
		r.mux.RLock()
		peer, ok := r.members[from]
		r.mux.RUnlock()

		if !ok {
			return nil
		}

		if peer.me.Role == AUDIENCE {
			return nil
		}

		// @TODO Rate limiting?

		go r.notify(&pb.SignalReply_Event{
			Type: pb.SignalReply_Event_LINK_SHARED,
			From: int64(from),
			Data: cmd.Data,
		})
	case pb.SignalRequest_Command_ADD_ADMIN:
		r.onAddAdmin(from, cmd)
	case pb.SignalRequest_Command_REMOVE_ADMIN:
		r.onRemoveAdmin(from, cmd)
	case pb.SignalRequest_Command_RENAME_ROOM:
		r.onRoomRename(from, cmd)
	}

	return nil
}

func (r *Room) onInvite(from int, invite *pb.Invite) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	peer := r.members[from]
	if peer == nil {
		return nil
	}

	if r.isPrivate && peer.me.Role != ADMIN {
		return nil
	}

	r.invited[int(invite.Id)] = true

	err := r.queue.Publish(pubsub.RoomTopic, pubsub.NewRoomInviteEvent(r.name, r.id, from, int(invite.Id)))
	if err != nil {
		log.Printf("queue.Publish err: %v\n", err)
	}

	return nil
}

func (r *Room) onKick(from int, kick *pb.Kick) error {
	if !r.isAdmin(from) {
		return nil
	}

	r.mux.Lock()
	defer r.mux.Unlock()
	r.kicked[int(kick.Id)] = true
	err := r.members[int(kick.Id)].rtc.Close()
	if err != nil {
		log.Printf("rtc.Close error %v\n", err)
	}

	return nil
}

func (r *Room) onAddAdmin(from int, add *pb.SignalRequest_Command) {
	if !r.isAdmin(from) {
		return
	}

	r.mux.Lock()
	r.members[bytesToInt(add.Data)].me.Role = ADMIN
	r.mux.Unlock()

	go r.notify(&pb.SignalReply_Event{
		Type: pb.SignalReply_Event_ADDED_ADMIN,
		From: int64(from),
		Data: add.Data,
	})
}

func (r *Room) onRemoveAdmin(from int, remove *pb.SignalRequest_Command) {
	if !r.isAdmin(from) {
		return
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	r.members[bytesToInt(remove.Data)].me.Role = SPEAKER

	go r.notify(&pb.SignalReply_Event{
		Type: pb.SignalReply_Event_REMOVED_ADMIN,
		From: int64(from),
		Data: remove.Data,
	})
}

func (r *Room) onRoomRename(from int, rename *pb.SignalRequest_Command) {
	if !r.isAdmin(from) {
		return
	}

	name := strings.TrimSpace(string(rename.Data))
	if len([]rune(name)) > 30 {
		name = string([]rune(name)[:30])
	}

	r.mux.Lock()
	r.name = name
	r.mux.Unlock()

	go r.notify(&pb.SignalReply_Event{
		Type: pb.SignalReply_Event_RENAMED_ROOM,
		From: int64(from),
		Data: rename.Data,
	})
}

func (r *Room) isAdmin(peer int) bool {
	r.mux.RLock()
	defer r.mux.RUnlock()

	p, ok := r.members[peer]
	if !ok {
		return false
	}

	return p.me.Role == ADMIN
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

	visibility := pb.Visibility_PUBLIC
	if r.isPrivate {
		visibility = pb.Visibility_PRIVATE
	}

	for _, member := range r.members {
		if member.me.SSRC == 0 {
			continue
		}

		members = append(members, &pb.RoomState_RoomMember{
			Id:          int64(member.me.ID),
			DisplayName: member.me.DisplayName,
			Image:       member.me.Image,
			Role:        string(member.me.Role),
			Muted:       member.me.IsMuted,
			Ssrc:        member.me.SSRC,
		})
	}

	state := &pb.RoomState{
		Id:         int64(r.id),
		Name:       r.name,
		Role:       string(SPEAKER), // @TODO THIS SHOULD DEPEND ON WHO OWNS THE ROOM ETC
		Members:    members,
		Visibility: visibility,
	}

	if r.group != nil {
		state.Group = &pb.RoomState_Group{
			Id:    int64(r.group.ID),
			Name:  r.group.Name,
			Image: r.group.Image,
		}
	}

	return state
}

func has(peers map[int]*peer, fn func(*peer) bool) bool {
	for _, peer := range peers {
		if fn(peer) {
			return true
		}
	}

	return false
}

func intToBytes(val int) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, uint64(val))
	return bytes
}

func bytesToInt(val []byte) int {
	return int(binary.LittleEndian.Uint64(val))
}
