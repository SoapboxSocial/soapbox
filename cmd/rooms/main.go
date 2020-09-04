package main

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
			},
		},
	}

	addr := ":50051"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Panicf("failed to listen: %v", err)
		return
	}

	s := grpc.NewServer()
	pb.RegisterRoomServiceServer(s, rooms.NewServer(sfu.NewSFU(config)))

	err = s.Serve(lis)
	if err != nil {
		log.Panicf("failed to serve: %v", err)
	}
}
