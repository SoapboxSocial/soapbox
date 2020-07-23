package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc"
)

var upgrader = websocket.Upgrader{} // use default options
type Msg struct {
	messageType int
	p []byte
}

type Payload struct {
	Payload map[string]interface{}
	Type string
}

var conns = make(map[net.Addr]chan Msg)

func main() {
	log.SetFlags(0)
	http.HandleFunc("/", test)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func test(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	out := make(chan Msg)

	conns[c.RemoteAddr()] = out

	defer func() {
		delete(conns, c.RemoteAddr())
		c.Close()
	}()

	go func() {
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}

			payload := &Payload{}
			err = json.Unmarshal(message, payload)
			if err != nil {
				fmt.Println(err)
				return
			}

			if payload.Type == "SessionDescription" {
				p := webrtc.SessionDescription{
					Type: getType(payload.Payload["type"].(string)),
					SDP: payload.Payload["sdp"].(string),
				}

				fmt.Println(p)
			}


			m := Msg{mt, message}

			for ip, peer := range conns {
				if ip == c.RemoteAddr() {
					continue
				}

				peer <- m
			}
		}
	}()

	for {
		msg := <- out
		err := c.WriteMessage(msg.messageType, msg.p)
		if err != nil {
			return
		}
	}
}

func getType(t string) webrtc.SDPType {
	switch t {
	case "offer":
		return webrtc.SDPTypeOffer
	case "prAnswer":
		return webrtc.SDPTypePranswer
	case "answer":
		return webrtc.SDPTypeAnswer
	}

	panic("fuck")
}