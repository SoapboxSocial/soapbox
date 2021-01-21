package rooms

import (
	"github.com/pion/ion-sfu/pkg/sfu"
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
