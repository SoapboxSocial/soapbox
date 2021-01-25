package rooms

import (
	"github.com/pion/ion-sfu/pkg/sfu"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

// @TODO ADD USER DATA

type Member struct {
	id   int
	peer *sfu.Peer
}

func (m *Member) Close() {
	_ = m.peer.Close()
}

func (m *Member) Notify(label string, data []byte) error {
	return m.peer.GetDataChannel(label).Send(data)
}

// @TODO
func (m *Member) ToProto() *pb.RoomState_RoomMember {
	return &pb.RoomState_RoomMember{
		Id: int64(m.id),
		DisplayName: "poop",
		Image: "",
		Muted: true,
		Role: pb.RoomState_RoomMember_REGULAR,
	}
}
