package rooms

import (
	"errors"
	"sync"
)

type Repository struct {
	mux sync.RWMutex

	rooms map[string]*Room
}

func NewRepository() *Repository {
	return &Repository{
		mux:         sync.RWMutex{},
		rooms: make(map[string]*Room),
	}
}

func (r *Repository) Set(room *Room) {
	r.mux.Lock()
	defer r.mux.Unlock()

	r.rooms[room.id] = room
}

func (r *Repository) Get(id string) (*Room, error) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	room, ok := r.rooms[id]
	if !ok {
		return nil, errors.New("room not found")
	}

	return room, nil
}
