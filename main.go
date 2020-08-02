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
	"github.com/pion/webrtc/v3"

	"github.com/ephemeral-networks/voicely/pkg/rooms"
)

type Member struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

type RoomPayload struct {
	ID      int      `json:"id"`
	Name    string   `json:"name,omitempty"`
	Members []Member `json:"members"`
}

type SDPPayload struct {
	Name *string `json:"name,omitempty"`
	ID   *int    `json:"id,omitempty"`
	SDP  string  `json:"sdp"`
	Type string  `json:"type"`
}

type JoinPayload struct {
	Name    string     `json:"name,omitempty"`
	Members []Member   `json:"members"`
	SDP     SDPPayload `json:"sdp"`
	Role    string     `json:"role"` // @todo find better name
}

// this represents a server session
type Session struct {
	User          int
	Token         string
}

type LoginState struct {
	Email string
	Pin   string
}

var tokens = make(map[string]LoginState)

type ErrorCode int

const (
	ErrorCodeRoomNotFound       ErrorCode = 1
	ErrorCodeRoomFailedToJoin             = 2
	ErrorCodeInvalidRequestBody           = 3
	ErrorCodeFailedToCreateRoom           = 4
)

func main() {

	manager := rooms.NewRoomManager()

	r := mux.NewRouter()

	r.HandleFunc("/v1/rooms", func(w http.ResponseWriter, r *http.Request) {
		data := make([]RoomPayload, 0)

		manager.MapRooms(func(room *rooms.Room) {
			if room == nil {
				return
			}

			r := RoomPayload{ID: room.GetID(), Members: make([]Member, 0)}

			name := room.GetName()
			if name != "" {
				r.Name = name
			}

			room.MapPeers(func(s string, peer rooms.Peer) {
				r.Members = append(r.Members, Member{s, string(peer.Role())})
			})

			data = append(data, r)
		})

		err := jsonEncode(w, data)
		if err != nil {
			fmt.Println(err)
		}
	}).Methods("GET")

	r.HandleFunc("/v1/rooms/create", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			jsonError(w, 400, ErrorCodeInvalidRequestBody, "invalid request body")
			return
		}

		payload := &SDPPayload{}
		err = json.Unmarshal(b, payload)
		if err != nil {
			jsonError(w, 400, ErrorCodeInvalidRequestBody, "invalid request body")
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

		name := ""
		if payload.Name != nil {
			name = *payload.Name
		}

		room := manager.CreateRoom(name)

		sdp, err := room.Join(r.RemoteAddr, p)
		if err != nil {
			manager.RemoveRoom(room.GetID())
			jsonError(w, 500, ErrorCodeFailedToCreateRoom, "failed to create room")
			return
		}

		id := room.GetID()
		resp := &SDPPayload{ID: &id, Type: strings.ToLower(sdp.Type.String()), SDP: sdp.SDP}

		err = jsonEncode(w, resp)
		if err != nil {
			manager.RemoveRoom(room.GetID())
			fmt.Println(err)
		}
	}).Methods("POST")

	r.HandleFunc("/v1/rooms/{id:[0-9]+}/join", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			jsonError(w, 400, ErrorCodeInvalidRequestBody, "invalid request body")
			return
		}

		payload := &SDPPayload{}
		err = json.Unmarshal(b, payload)
		if err != nil {
			jsonError(w, 400, ErrorCodeInvalidRequestBody, "invalid request body")
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
			jsonError(w, 404, ErrorCodeRoomNotFound, "room not found")
			return
		}

		sdp, err := room.Join(r.RemoteAddr, p)
		if err != nil {
			jsonError(w, 500, ErrorCodeRoomFailedToJoin, "failed to join room")
			return
		}

		members := make([]Member, 0)

		room.MapPeers(func(s string, peer rooms.Peer) {
			// @todo will need changing
			if s == r.RemoteAddr {
				return
			}

			members = append(members, Member{s, string(peer.Role())})
		})

		resp := &JoinPayload{
			Members: members,
			SDP: SDPPayload{
				ID:   &id,
				Type: strings.ToLower(sdp.Type.String()),
				SDP:  sdp.SDP,
			},
			Role: string(room.GetRoleForPeer(r.RemoteAddr)),
		}

		name := room.GetName()
		if name != "" {
			resp.Name = name
		}

		err = jsonEncode(w, resp)
		if err != nil {
			fmt.Println(err)
		}
	}).Methods("POST")

	// @todo so this is how login will work:
	//   - users submits email
	//   - check if exists, generate token, send pin
	//   - submit received email pin
	//   - if pin match, login

	r.HandleFunc("/v1/login/start", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			// @todo
			fmt.Println("fuck")
			return
		}
		email := r.Form.Get("email")
		if email == "" {
			// @todo
			return
		}

		// @todo check that email is set
		token := GenerateToken()
		pin := GeneratePin()

		tokens[token] = LoginState{Email: email, Pin: pin}

		// @todo cleanup
		err = json.NewEncoder(w).Encode(map[string]string{"token": token})
		if err != nil {
			fmt.Println(err)
		}

		log.Println("pin:" + pin)

	}).Methods("POST")

	r.HandleFunc("/v1/login/pin", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			// @todo
			fmt.Println("fuck")
			return
		}

		token := r.Form.Get("token")
		pin := r.Form.Get("pin")

		state := tokens[token]
		if state.Pin != pin {
			// @todo send failure
			return
		}

		fmt.Println(state)

		log.Println("success")

		// @todo make account if not exist

		// @todo start session
	}).Methods("POST")

	r.HandleFunc("/v1/login/register", func(writer http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			// @todo
			fmt.Println("fuck")
			return
		}

		//token := r.Form.Get("token")
	})

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

func jsonError(w http.ResponseWriter, responseCode int, code ErrorCode, msg string) {
	type ErrorResponse struct {
		Code    ErrorCode `json:"code"`
		Message string    `json:"message"`
	}

	resp, err := json.Marshal(ErrorResponse{Code: code, Message: msg})
	if err != nil {
		log.Println("failed encoding error")
		return
	}

	w.WriteHeader(responseCode)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("failed to encode response: %s", err.Error())
	}
}

func jsonEncode(w http.ResponseWriter, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
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
