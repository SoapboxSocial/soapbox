package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

var list = &cobra.Command{
	Use:   "list",
	Short: "list all active rooms",
	RunE:  runList,
}

func runList(*cobra.Command, []string) error {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	client := pb.NewRoomServiceClient(conn)

	resp, err := client.ListRooms(context.TODO(), &pb.ListRoomsRequest{})
	if err != nil {
		return err
	}

	for _, room := range resp.Rooms {
		log.Printf("Room (%s) Peers = %d Visibility = %s", room.Id, len(room.Members), room.Visibility)
	}

	return nil
}
