package signalling

import (
	"github.com/gorilla/websocket"
)

type Packet struct {

}

type Connection struct {
	conn *websocket.Conn
	send chan Packet
}

func NewConnection(conn *websocket.Conn) *Connection {
	c := &Connection{
		conn: conn,
		send: make(chan Packet),
	}

	return c
}
