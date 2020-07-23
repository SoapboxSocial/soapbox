package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/pion/webrtc"

	"github.com/ephemeral-networks/voicely/pkg/rooms"
)

type Payload struct {
	Payload map[string]interface{} `json:"payload"`
	Type    string                 `json:"type"`
}

func main() {
	room := rooms.NewRoom()

	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ok")
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		payload := &Payload{}
		err = json.Unmarshal(b, payload)
		if err != nil {
			// @todo
			return
		}

		if payload.Type != "SessionDescription" {
			// @todo
			return
		}

		err, t := getType(payload.Payload["type"].(string))
		if err != nil {
			// @todo more errors this shit is invalid
			return
		}

		p := webrtc.SessionDescription{
			Type: t,
			SDP:  payload.Payload["sdp"].(string),
		}

		err, resp := room.Join(r.RemoteAddr, p)
		if err != nil {
			// @todo handle
			return
		}

		pack := Payload{
			Type:    "SessionDescription",
			Payload: make(map[string]interface{}),
		}

		pack.Payload["sdp"] = resp.SDP
		pack.Payload["type"] = strings.ToLower(resp.Type.String())

		err = json.NewEncoder(w).Encode(pack)
		if err != nil {
			fmt.Println(err)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getType(t string) (error, webrtc.SDPType) {
	switch t {
	case "offer":
		return nil, webrtc.SDPTypeOffer
	case "prAnswer":
		return nil, webrtc.SDPTypePranswer
	case "answer":
		return nil, webrtc.SDPTypeAnswer
	}

	return fmt.Errorf("unknown type: %s", t), webrtc.SDPType(-1)
}
