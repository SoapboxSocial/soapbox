package rooms

import (
	"github.com/pion/ion-sfu/pkg/sfu"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type Member struct {
	id   int
	peer *sfu.Peer
}

func (m *Member) Receive() (*pb.Command, error) {
	return nil, nil
}

func (m *Member) Notify(event *pb.Event) {

}
