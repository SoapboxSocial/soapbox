package rooms

type RoomManger struct {
	rooms []*Room
}

func (rm *RoomManger) CreateRoom() *Room {
	r := &Room{}

	//go r.run()

	return r
}

