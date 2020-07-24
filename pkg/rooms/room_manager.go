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

// @todo this will probably be very inefficient at scale lol
func (rm *RoomManger) MapRooms(fn func(*Room)) {
	rm.RLock()
	defer rm.RUnlock()

	for _, r := range rm.rooms {
		fn(r)
	}
}

func (rm *RoomManger) CreateRoom() *Room {
	rm.Lock()
	defer rm.Unlock()

	r := NewRoom(len(rm.rooms))
	rm.rooms = append(rm.rooms, r)

	return r
}

