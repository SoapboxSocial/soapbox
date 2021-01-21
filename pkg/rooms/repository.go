package rooms

import (
	"errors"
	"sync"
)

type RoomRepository struct {
	mux sync.RWMutex

	rooms map[string]*Room
}

func (r *RoomRepository) Set(room *Room) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.rooms[room.id] = room
}

func (r *RoomRepository) Get(id string) (*Room, error) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	room, ok := r.rooms[id]
	if !ok {
		return nil, errors.New("room not found")
	}

	return room, nil
}
