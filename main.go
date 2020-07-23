package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pion/webrtc"

	"github.com/ephemeral-networks/voicely/pkg/rooms"
)

type SDPPayload struct {
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

func main() {
	room := rooms.NewRoom()

	r := mux.NewRouter()


	r.HandleFunc("/v1/rooms/{id:[0-9]+}/join", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		payload := &SDPPayload{}
		err = json.Unmarshal(b, payload)
		if err != nil {
			log.Printf("failed to decode payload: %s\n", err.Error())
			return
		}

		err, t := getType(payload.Type)
		if err != nil {
			// @todo more errors this shit is invalid
			return
		}

		p := webrtc.SessionDescription{
			Type: t,
			SDP:  payload.SDP,
		}

		err, sdp := room.Join(r.RemoteAddr, p)
		if err != nil {
			// @todo handle
			return
		}

		resp := SDPPayload{
			Type: strings.ToLower(sdp.Type.String()),
			SDP:  sdp.SDP,
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			fmt.Println(err)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", r))
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
