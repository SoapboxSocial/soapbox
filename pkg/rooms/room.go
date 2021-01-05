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

func (r *Room) onMessage(command *pb.Command) {
	switch command.Payload.(type) {
	case *pb.Command_Mute_:
		break
	case *pb.Command_Unmute_:
		break
	case *pb.Command_Reaction_:
		break
	case *pb.Command_LinkShare_:
		break
	case *pb.Command_InviteAdmin_:
		break
	case *pb.Command_AcceptAdmin_:
		break
	case *pb.Command_RemoveAdmin_:
		break
	case *pb.Command_RenameRoom_:
		break
	case *pb.Command_InviteUser_:
		break
	case *pb.Command_KickUser_:
		break
	case *pb.Command_MuteUser_:
		break
	case *pb.Command_RecordScreen_:
		break
	}
}
