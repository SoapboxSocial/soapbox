package rooms

import (
	"log"
	"net"

	sfu "github.com/pion/ion-sfu/pkg"
	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

func main() {

	config := sfu.Config{
		WebRTC: sfu.WebRTCConfig{
			ICEPortRange: []uint16{1000, 6000},
			ICEServers: []sfu.ICEServerConfig{
				{
					URLs: []string{
						"stun:stun.l.google.com:19302",
						"stun:stun1.l.google.com:19302",
						"stun:stun2.l.google.com:19302",
						"stun:stun3.l.google.com:19302",
						"stun:stun4.l.google.com:19302",
					},
				},
				{
					URLs:       []string{"turn:turn.awsome.org:3478"},
					Username:   "awsome",
					Credential: "awsome",
				},
			},
		},
		Receiver: sfu.ReceiverConfig{
			Video: sfu.WebRTCVideoReceiverConfig{
				REMBCycle:     2,
				PLICycle:      1,
				TCCCycle:      1,
				MaxBandwidth:  1000,
				MaxBufferTime: 5000,
			},
		},
	}

	addr := ":50051"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterRoomServiceServer(s, rooms.NewServer(sfu.NewSFU(config)))
	if err := s.Serve(lis); err != nil {
		log.Panicf("failed to serve: %v", err)
	}

	select {}
}
