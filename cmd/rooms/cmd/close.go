package cmd

import (
	"context"
	"errors"
	"fmt"

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
	close.Flags().StringVarP(&addr, "addr", "a", "127.0.0.1:50052", "grpc address")
}

func runClose(*cobra.Command, []string) error {
	if room == "" {
		return errors.New("room is empty")
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	defer conn.Close()

	client := pb.NewRoomServiceClient(conn)

	resp, err := client.CloseRoom(context.TODO(), &pb.CloseRoomRequest{Id: room})
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New("failed to close room")
	}

	fmt.Printf("Closed %s\n", room)

	return nil
}
