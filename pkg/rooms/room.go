package rooms

import (
	"sync"

	"github.com/pion/ion-sfu/pkg/sfu"
)

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
