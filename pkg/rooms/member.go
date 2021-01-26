package rooms

import (
	"sync"

	"github.com/pion/ion-sfu/pkg/sfu"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type Member struct {
	mux sync.RWMutex

	id    int
	name  string
	image string
	muted bool
	role  pb.RoomState_RoomMember_Role

	peer *sfu.Peer
}

func NewMember(id int, name, image string, peer *sfu.Peer) *Member {
	return &Member{
		id:    id,
		name:  name,
		image: image,
		muted: true,
		peer:  peer,
		role:  pb.RoomState_RoomMember_REGULAR,
	}
}

func (m *Member) Mute() {
	m.mux.Lock()
	defer m.mux.RUnlock()

	m.muted = true
}

func (m *Member) Unmute() {
	m.mux.Lock()
	defer m.mux.RUnlock()

	m.muted = false
}

func (m *Member) SetRole(role pb.RoomState_RoomMember_Role) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.role = role
}

func (m *Member) Close() {
	_ = m.peer.Close()
}

func (m *Member) Notify(label string, data []byte) error {
	return m.peer.GetDataChannel(label).Send(data)
}

func (m *Member) Role() pb.RoomState_RoomMember_Role {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return m.role
}

func (m *Member) ToProto() *pb.RoomState_RoomMember {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return &pb.RoomState_RoomMember{
		Id:          int64(m.id),
		DisplayName: m.name,
		Image:       m.image,
		Muted:       m.muted,
		Role:        m.role,
	}
}
