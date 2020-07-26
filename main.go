package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pion/webrtc"

	"github.com/ephemeral-networks/voicely/pkg/rooms"
)

type SDPPayload struct {
	ID   *int    `json:"id,omitempty"`
	SDP  string `json:"sdp"`
	Type string `json:"type"`
}

func main() {

	manager := rooms.NewRoomManager()

	r := mux.NewRouter()

	r.HandleFunc("/v1/rooms", func(w http.ResponseWriter, r *http.Request) {
		data := make([]int, 0)

		manager.MapRooms(func(room *rooms.Room) {
			if room == nil {
				return
			}

			data = append(data, room.GetID())
		})

		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			fmt.Println(err)
		}
	}).Methods("GET")

	r.HandleFunc("/v1/rooms/create", func(w http.ResponseWriter, r *http.Request) {
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

		room := manager.CreateRoom()

		sdp, err := room.Join(r.RemoteAddr, p)
		if err != nil {
			// @todo
			return
		}

		id := room.GetID()

		resp := SDPPayload{
			ID:   &id,
			Type: strings.ToLower(sdp.Type.String()),
			SDP:  sdp.SDP,
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			fmt.Println(err)
		}
	}).Methods("POST")

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

		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			return
		}

		room, err := manager.GetRoom(id)
		if err != nil {
			// @todo handle
			return
		}

		sdp, err := room.Join(r.RemoteAddr, p)
		if err != nil {
			// @todo
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
	}).Methods("POST")

	// @todo so this is how login will work:
	//   - users submits email
	//   - check if exists, generate token, send pin
	//   - submit received email pin
	//   - if pin match, login

	var pins = make(map[string]string)

	r.HandleFunc("/v1/login/start", func(w http.ResponseWriter, r *http.Request) {
		email := r.Form.Get("email")
		fmt.Println(email)

		// @todo check that email is set
		token := GenerateToken()
		pin := GeneratePin()

		pins[token] = pin

		// @todo cleanup
		err := json.NewEncoder(w).Encode(map[string]string{"token": token})
		if err != nil {
			fmt.Println(err)
		}

	}).Methods("POST")

	r.HandleFunc("/v1/login/pin", func(w http.ResponseWriter, r *http.Request) {
		token := r.Form.Get("token")
		pin := r.Form.Get("pin")

		if pins[token] != pin {
			// @todo send failure
			return
		}

		// @todo start session
	}).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", r))
}

func GenerateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func GeneratePin() string {
	max := 6
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

var table = []byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

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
