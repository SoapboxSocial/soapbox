package grpc_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/rooms"
	"github.com/soapboxsocial/soapbox/pkg/rooms/grpc"
	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

func TestService_RegisterWelcomeRoom(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	repository := rooms.NewRepository()
	ws := rooms.NewWelcomeStore(rdb)

	service := grpc.NewService(repository, ws, nil)

	userID := int64(1)
	resp, err := service.RegisterWelcomeRoom(context.Background(), &pb.RegisterWelcomeRoomRequest{UserId: userID})
	if err != nil {
		t.Fatal(err)
	}

	id, err := ws.GetUserIDForWelcomeRoom(resp.Id)
	if err != nil {
		t.Fatal(err)
	}

	if int64(id) != userID {
		t.Errorf("%d does not equal %d", id, userID)
	}
}
