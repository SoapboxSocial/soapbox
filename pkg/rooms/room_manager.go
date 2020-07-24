package rooms

import (
	"fmt"
	"sync"
)

type RoomManger struct {
	sync.RWMutex

	// @todo in the future this will need to be a map with IDs probably
	rooms []*Room
}

func NewRoomManager() *RoomManger {
	return &RoomManger{
		rooms: make([]*Room, 0),
	}
}

func (rm *RoomManger) GetRoom(i int) (*Room, error) {
	rm.RLock()
	defer rm.RUnlock()

	if len(rm.rooms) <= i {
		return nil, fmt.Errorf("room %d does not exist", i)
	}

	return rm.rooms[i], nil
}

func (rm *RoomManger) CreateRoom() *Room {
	rm.Lock()
	defer rm.Unlock()

	r := NewRoom(len(rm.rooms))
	rm.rooms = append(rm.rooms, r)

	return r
}

