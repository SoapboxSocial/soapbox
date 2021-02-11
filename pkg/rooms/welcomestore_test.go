package rooms_test

import (
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/rooms"
)

func TestWelcomeStore(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	ws := rooms.NewWelcomeStore(rdb)

	room := "roomID-123"
	user := int64(1)

	err = ws.StoreWelcomeRoomID(room, user)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := ws.GetUserIDForWelcomeRoom(room)
	if err != nil {
		t.Fatal(err)
	}

	if int64(resp) != user {
		t.Errorf("%d did not match %d", resp, user)
	}

	err = ws.DeleteWelcomeRoom(room)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ws.GetUserIDForWelcomeRoom(room)
	if err == nil {
		t.Error("unexpected return value")
	}
}
