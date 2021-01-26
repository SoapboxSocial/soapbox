package rooms

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/pion/ion-sfu/pkg/sfu"
	"github.com/pion/webrtc/v3"
	"google.golang.org/protobuf/proto"

	"github.com/soapboxsocial/soapbox/pkg/rooms/internal"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

const CHANNEL = "soapbox"

type Room struct {
	mux sync.RWMutex

	id   string
	name string

	members map[int]*Member

	adminInvites map[int]bool
	kicked       map[int]bool
	invited      map[int]bool

	peerToMember map[string]int

	onDisconnectedHandlerFunc func(room string, id int)

	session *sfu.Session
}

func NewRoom(id, name string, session *sfu.Session) *Room {
	r := &Room{
		id:           id,
		name:         name,
		session:      session,
		peerToMember: make(map[string]int),
		members:      make(map[int]*Member),
	}

	dc := sfu.NewDataChannel(CHANNEL)
	dc.OnMessage(func(ctx context.Context, args sfu.ProcessArgs) {
		m := &pb.Command{}
		err := proto.Unmarshal(args.Message.Data, m)
		if err != nil {
			log.Printf("error unmarshalling: %v", err)
			return
		}

		r.mux.RLock()
		defer r.mux.RUnlock()

		user, ok := r.peerToMember[args.Peer.ID()]
		if !ok {
			return
		}

		r.onMessage(user, m)
	})

	session.AddDataChannel(dc)

	return r
}

func (r *Room) ID() string {
	return r.id
}

func (r *Room) PeerCount() int {
	r.mux.RLock()
	defer r.mux.RUnlock()
	return len(r.members)
}

func (r *Room) IsAdmin(id int) bool {
	r.mux.RLock()
	defer r.mux.RUnlock()

	member, ok := r.members[id]
	if !ok {
		return false
	}

	return member.Role() == pb.RoomState_RoomMember_ADMIN
}

func (r *Room) MapMembers(f func(member *Member)) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	for _, member := range r.members {
		f(member)
	}
}

func (r *Room) OnDisconnected(f func(room string, id int)) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.onDisconnectedHandlerFunc = f
}

func (r *Room) ToProtoForPeer() *pb.RoomState {
	r.mux.RLock()
	defer r.mux.RUnlock()

	members := make([]*pb.RoomState_RoomMember, 0)

	for _, member := range r.members {
		members = append(members, member.ToProto())
	}

	return &pb.RoomState{
		Id:      r.id,
		Name:    r.name,
		Members: members,
	}
}

func (r *Room) Handle(user *users.User, peer *sfu.Peer) {
	r.peerToMember[peer.ID()] = user.ID

	me := NewMember(user.ID, user.DisplayName, user.Image, peer)

	r.mux.Lock()
	r.members[user.ID] = me
	r.mux.Unlock()
	peer.OnICEConnectionStateChange = func(state webrtc.ICEConnectionState) {
		log.Printf("connection state changed %d", state)

		switch state {
		case webrtc.ICEConnectionStateConnected:
			r.notify(&pb.Event{
				From: int64(user.ID),
				Payload: &pb.Event_Joined_{
					Joined: &pb.Event_Joined{User: me.ToProto()},
				},
			})

			dc := peer.GetDataChannel(CHANNEL)
			if dc == nil {
				fmt.Println("data channel not found")
				return
			}

			dc.OnClose(func() {
				r.onDisconnected(int64(user.ID))
			})
		case webrtc.ICEConnectionStateDisconnected:
			r.onDisconnected(int64(user.ID))
		}
	}
}

func (r *Room) onDisconnected(id int64) {
	log.Printf("disconnected %d", id)

	r.mux.RLock()
	peer, ok := r.members[int(id)]
	r.mux.RUnlock()

	if !ok {
		return
	}

	r.notify(&pb.Event{
		From:    id,
		Payload: &pb.Event_Left_{},
	})

	err := peer.Close()
	if err != nil {
		log.Printf("rtc.Close error %v\n", err)
	}

	r.mux.Lock()
	delete(r.members, int(id))
	r.mux.Unlock()

	// @TODO NEW ADMIN

	r.onDisconnectedHandlerFunc(r.id, int(id))
}

func (r *Room) onMessage(from int, command *pb.Command) {
	switch command.Payload.(type) {
	case *pb.Command_MuteUpdate_:
		r.onMuteUpdate(from, command.GetMuteUpdate())
	case *pb.Command_Reaction_:
		r.onReaction(from, command.GetReaction())
	case *pb.Command_LinkShare_:
		r.onLinkShare(from, command.GetLinkShare())
	case *pb.Command_InviteAdmin_:
		r.onInviteAdmin(from, command.GetInviteAdmin())
	case *pb.Command_AcceptAdmin_:
		r.onAcceptAdmin(from)
	case *pb.Command_RemoveAdmin_:
		r.onRemoveAdmin(from, command.GetRemoveAdmin())
	case *pb.Command_RenameRoom_:
		r.onRenameRoom(from, command.GetRenameRoom())
	case *pb.Command_InviteUser_:
		r.onInviteUser(from, command.GetInviteUser())
	case *pb.Command_KickUser_:
		r.onKickUser(from, command.GetKickUser())
	case *pb.Command_MuteUser_:
		r.onMuteUser(from, command.GetMuteUser())
	case *pb.Command_RecordScreen_:
		r.onRecordScreen(from)
	}
}

func (r *Room) onMuteUpdate(from int, cmd *pb.Command_MuteUpdate) {
	member := r.member(from)
	if member == nil {
		log.Printf("member %d not found", from)
		return
	}

	if cmd.Muted {
		member.Mute()
	} else {
		member.Unmute()
	}

	r.notify(&pb.Event{
		From:    int64(from),
		Payload: &pb.Event_MuteUpdated_{MuteUpdated: &pb.Event_MuteUpdated{IsMuted: cmd.Muted}},
	})
}

func (r *Room) onReaction(from int, cmd *pb.Command_Reaction) {
	r.notify(&pb.Event{
		From:    int64(from),
		Payload: &pb.Event_Reacted_{Reacted: &pb.Event_Reacted{Emoji: cmd.Emoji}},
	})
}

func (r *Room) onLinkShare(from int, cmd *pb.Command_LinkShare) {
	r.notify(&pb.Event{
		From:    int64(from),
		Payload: &pb.Event_LinkShared_{LinkShared: &pb.Event_LinkShared{Link: cmd.Link}},
	})
}

// @TODO
func (r *Room) onInviteAdmin(from int, cmd *pb.Command_InviteAdmin) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if !r.IsAdmin(from) {
		return
	}

	member, ok := r.members[int(cmd.Id)]
	if !ok {
		return
	}

	r.adminInvites[int(cmd.Id)] = true

	data, err := proto.Marshal(cmd)
	if err != nil {
		log.Printf("failed to marshal %v", err)
		return
	}

	member.Notify(CHANNEL, data)
}

func (r *Room) onAcceptAdmin(from int) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if !r.adminInvites[from] {
		return
	}

	delete(r.adminInvites, from)

	member, ok := r.members[from]
	if !ok {
		return
	}

	member.SetRole(pb.RoomState_RoomMember_ADMIN)

	r.notify(&pb.Event{
		From:    int64(from),
		Payload: &pb.Event_AddedAdmin_{AddedAdmin: &pb.Event_AddedAdmin{}},
	})
}

func (r *Room) onRemoveAdmin(from int, cmd *pb.Command_RemoveAdmin) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if !r.IsAdmin(from) {
		return
	}

	member := r.member(from)
	if member == nil {
		log.Printf("member %d not found", from)
		return
	}

	member.SetRole(pb.RoomState_RoomMember_ADMIN)

	r.notify(&pb.Event{
		From:    int64(from),
		Payload: &pb.Event_RemovedAdmin_{RemovedAdmin: &pb.Event_RemovedAdmin{Id: cmd.Id}},
	})
}

func (r *Room) onRenameRoom(from int, cmd *pb.Command_RenameRoom) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if !r.IsAdmin(from) {
		return
	}

	r.name = internal.TrimRoomNameToLimit(cmd.Name)

	r.notify(&pb.Event{
		From:    int64(from),
		Payload: &pb.Event_RenamedRoom_{RenamedRoom: &pb.Event_RenamedRoom{Name: r.name}},
	})
}

func (r *Room) onInviteUser(from int, cmd *pb.Command_InviteUser) {

}

func (r *Room) onKickUser(from int, cmd *pb.Command_KickUser) {
	if !r.IsAdmin(from) {
		return
	}

	p, ok := r.members[int(cmd.Id)]
	if !ok {
		return
	}

	_ = p.Close()
}

func (r *Room) onMuteUser(from int, cmd *pb.Command_MuteUser) {

}

func (r *Room) onRecordScreen(from int) {
	r.notify(&pb.Event{
		From:    int64(from),
		Payload: &pb.Event_RecordedScreen_{RecordedScreen: &pb.Event_RecordedScreen{}},
	})
}

func (r *Room) member(id int) *Member {
	r.mux.RLock()
	defer r.mux.RUnlock()

	member, ok := r.members[id]
	if !ok {
		return nil
	}

	return member
}

func (r *Room) notify(event *pb.Event) {
	data, err := proto.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal: %v", err)
		return
	}

	r.mux.RLock()
	defer r.mux.RUnlock()

	for id, member := range r.members {
		if id == int(event.From) {
			continue
		}

		log.Printf("notify %d", id)

		err := member.Notify(CHANNEL, data)
		if err != nil {
			log.Printf("failed to notify: %v\n", err)
		}
	}
}
