package rooms

import (
	"sync"

	"github.com/pion/ion-sfu/pkg/sfu"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type Member struct {
	id   int
	peer *sfu.Peer
}

type Room struct {
	mux sync.RWMutex

	id string
}

func (r *Room) ID() string {
	return r.id
}

func (r *Room) PeerCount() int {
	return 0
}

func (r *Room) Handle(user int, peer *sfu.Peer) error {
	return nil
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

}

func (r *Room) onUnmute(from int) {

}

func (r *Room) onReaction(from int, cmd *pb.Command_Reaction) {

}

func (r *Room) onLinkShare(from int, cmd *pb.Command_LinkShare) {

}

func (r *Room) onInviteAdmin(from int, cmd *pb.Command_InviteAdmin) {

}

func (r *Room) onAcceptAdmin(from int) {

}

func (r *Room) onRemoveAdmin(from int, cmd *pb.Command_RemoveAdmin) {

}

func (r *Room) onRenameRoom(from int, cmd *pb.Command_RenameRoom) {

}

func (r *Room) onInviteUser(from int, cmd *pb.Command_InviteUser) {

}

func (r *Room) onKickUser(from int, cmd *pb.Command_KickUser) {

}

func (r *Room) onMuteUser(from int, cmd *pb.Command_MuteUser) {

}

func (r *Room) onRecordScreen(from int) {

}

