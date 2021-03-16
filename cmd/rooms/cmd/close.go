package cmd

import (
	"context"
	"errors"
	"log"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

var close = &cobra.Command{
	Use:   "close",
	Short: "close an active room",
	RunE:  runClose,
}

var room string

func init() {
	close.Flags().StringVarP(&room, "room", "r", "", "room id")
}

func runClose(*cobra.Command, []string) error {
	if room == "" {
		return errors.New("room is empty")
	}

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
