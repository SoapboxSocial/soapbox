package signal

import (
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"

	"github.com/soapboxsocial/soapbox/pkg/rooms/pb"
)

type Transport interface {
	ReadMsg() (*pb.SignalRequest, error)
	Write(msg *pb.SignalReply) error
	Close() error
}

type WebSocketTransport struct {
	conn *websocket.Conn
}

func NewWebSocketTransport(conn *websocket.Conn) *WebSocketTransport {
	return &WebSocketTransport{
		conn: conn,
	}
}

func (w *WebSocketTransport) ReadMsg() (*pb.SignalRequest, error) {
	_, data, err := w.conn.ReadMessage()
	if err != nil {
		return nil, err
	}

	msg := &pb.SignalRequest{}
	err = proto.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (w *WebSocketTransport) Write(msg *pb.SignalReply) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	return w.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (w *WebSocketTransport) Close() error {
	return w.conn.Close()
}
