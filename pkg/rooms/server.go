package rooms

import (
	"io"
	"log"
	"sync"

	"github.com/pion/ion-sfu/pkg/sfu"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/soapboxsocial/soapbox/pkg/groups"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Server struct {
	mux sync.RWMutex

	sfu    *sfu.SFU
	sm     *sessions.SessionManager
	ub     *users.UserBackend
	groups *groups.Backend
	queue  *pubsub.Queue

	currentRoom *CurrentRoomBackend

	rooms map[string]*Room
}

func (s *Server) Signal(stream pb.SFU_SignalServer) error {
	peer := sfu.NewPeer(s.sfu)
	for {
		in, err := stream.Recv()
		if err != nil {
			_ = peer.Close()

			if err == io.EOF {
				return nil
			}

			errStatus, _ := status.FromError(err)
			if errStatus.Code() == codes.Canceled {
				return nil
			}

			log.Printf("signal error %v %v", errStatus.Message(), errStatus.Code())
			return err
		}
	}
}