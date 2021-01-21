package rooms

import (
	"sync"

	"github.com/pion/ion-sfu/pkg/sfu"

	"github.com/soapboxsocial/soapbox/pkg/rooms/internal"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type Room struct {
	mux sync.RWMutex

	id   string
	name string

	members map[int]Member

	adminInvites map[int]bool
	kicked       map[int]bool
	invited      map[int]bool
}

func NewRoom(id, name string) *Room {
	return &Room{id: id, name: name}
}

func (r *Room) ID() string {
	return r.id
}

func (r *Room) PeerCount() int {
	return 0
}

func (r *Room) IsAdmin(id int) bool {
	return false // @TODO
}

func (r *Room) ToProtoForPeer() *pb.RoomState {
	r.mux.RLock()
	defer r.mux.RUnlock()

	return &pb.RoomState{
		Id:   r.id,
		Name: r.name,
	}
}

func (r *Room) Handle(user int, peer *sfu.Peer) error {
	for {

	}
}

func (r *Room) onMessage(from int, command *pb.Command) {
	switch command.Payload.(type) {
	case *pb.Command_Mute_:
		r.onMute(from)
	case *pb.Command_Unmute_:
		r.onUnmute(from)
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

func (r *Room) onMute(from int) {

	// @TODO MARK MUTED

	r.notify(&pb.Event{
		From:    int64(from),
		Payload: &pb.Event_Muted_{Muted: &pb.Event_Muted{}},
	})
}

func (r *Room) onUnmute(from int) {

	// @TODO MARK UNMUTED

	r.notify(&pb.Event{
		From:    int64(from),
		Payload: &pb.Event_Unmuted_{Unmuted: &pb.Event_Unmuted{}},
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

	member.Notify(&pb.Event{
		From:    int64(from),
		Payload: &pb.Event_InvitedAdmin_{InvitedAdmin: &pb.Event_InvitedAdmin{}},
	})
}

func (r *Room) onAcceptAdmin(from int) {
	r.mux.Lock()
	defer r.mux.Unlock()

	if !r.adminInvites[from] {
		return
	}

	delete(r.adminInvites, from)

	_, ok := r.members[from]
	if !ok {
		return
	}

	// @TODO MARK ADMIN

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

	// @TODO
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

}

func (r *Room) onMuteUser(from int, cmd *pb.Command_MuteUser) {

}

func (r *Room) onRecordScreen(from int) {
	r.notify(&pb.Event{
		From:    int64(from),
		Payload: &pb.Event_RecordedScreen_{RecordedScreen: &pb.Event_RecordedScreen{}},
	})
}

func (r *Room) notify(event *pb.Event) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	for id, member := range r.members {
		if id == int(event.From) {
			continue
		}

		member.Notify(event)
	}
}
