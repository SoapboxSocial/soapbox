package rooms

import (
	"fmt"
	"log"
	"sync"
)

type RoomManger struct {
	sync.RWMutex

	// @todo in the future this will need to be a map with IDs probably
	rooms map[int]*Room

	nextID int
}

func NewRoomManager() *RoomManger {
	return &RoomManger{
		rooms:  make(map[int]*Room, 0),
		nextID: 0,
	}
}

func (rm *RoomManger) GetRoom(i int) (*Room, error) {
	rm.RLock()
	defer rm.RUnlock()

	r, ok := rm.rooms[i]
	if !ok {
		return nil, fmt.Errorf("room %d does not exist", i)
	}

	return r, nil
}

// @todo this will probably be very inefficient at scale lol
func (rm *RoomManger) MapRooms(fn func(*Room)) {
	rm.RLock()
	defer rm.RUnlock()

	for _, r := range rm.rooms {
		fn(r)
	}
}

func (rm *RoomManger) CreateRoom(name string) *Room {
	rm.Lock()
	defer rm.Unlock()

	listener := make(chan bool)

	id := rm.nextID
	r := NewRoom(id, name, listener)

	rm.rooms[id] = r
	rm.nextID++

	log.Printf("room %d created - current room count: %d\n", id, len(rm.rooms))

	go func() {
		for {
			<-listener

			// @todo if the owner left, elect a new one

			if r.PeerCount() == 0 {
				rm.Lock()
				delete(rm.rooms, id)
				rm.Unlock()

				log.Printf("room %d closed - current room count: %d\n", id, len(rm.rooms))
			}
		}
	}()

	return r
}

func (rm *RoomManger) RemoveRoom(id int) {
	rm.Lock()
	defer rm.Unlock()
	delete(rm.rooms, id)
}
