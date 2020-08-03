package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"

	"github.com/gorilla/mux"
	"github.com/pion/webrtc/v3"

	"github.com/ephemeral-networks/voicely/pkg/api/login"
	httputil "github.com/ephemeral-networks/voicely/pkg/http"
	"github.com/ephemeral-networks/voicely/pkg/rooms"
	"github.com/ephemeral-networks/voicely/pkg/sessions"
	"github.com/ephemeral-networks/voicely/pkg/users"
)

type Member struct {
	ID      string `json:"id"`
	Role    string `json:"role"`
	IsMuted bool   `json:"is_muted"`
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

func main() {
	db, err := sql.Open("postgres","host=127.0.0.1 port=5432 user=voicely password=voicely dbname=voicely sslmode=disable")
	if err != nil {
		panic(err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	s := sessions.NewSessionManager(rdb)
	ub := users.NewUserBackend(db)
	loginHandlers := login.NewLoginEndpoint(ub, s)

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
				r.Members = append(r.Members, Member{s, string(peer.Role()), peer.IsMuted()})
			})

			data = append(data, r)
		})

		err := httputil.JsonEncode(w, data)
		if err != nil {
			fmt.Println(err)
		}
	}).Methods("GET")

	r.HandleFunc("/v1/rooms/create", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid request body")
			return
		}

		payload := &SDPPayload{}
		err = json.Unmarshal(b, payload)
		if err != nil {
			httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid request body")
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
			httputil.JsonError(w, 500, httputil.ErrorCodeFailedToCreateRoom, "failed to create room")
			return
		}

		id := room.GetID()
		resp := &SDPPayload{ID: &id, Type: strings.ToLower(sdp.Type.String()), SDP: sdp.SDP}

		err = httputil.JsonEncode(w, resp)
		if err != nil {
			manager.RemoveRoom(room.GetID())
			fmt.Println(err)
		}
	}).Methods("POST")

	r.HandleFunc("/v1/rooms/{id:[0-9]+}/join", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid request body")
			return
		}

		payload := &SDPPayload{}
		err = json.Unmarshal(b, payload)
		if err != nil {
			httputil.JsonError(w, 400, httputil.ErrorCodeInvalidRequestBody, "invalid request body")
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
			httputil.JsonError(w, 404, httputil.ErrorCodeRoomNotFound, "room not found")
			return
		}

		sdp, err := room.Join(r.RemoteAddr, p)
		if err != nil {
			httputil.JsonError(w, 500, httputil.ErrorCodeRoomFailedToJoin, "failed to join room")
			return
		}

		members := make([]Member, 0)

		room.MapPeers(func(s string, peer rooms.Peer) {
			// @todo will need changing
			if s == r.RemoteAddr {
				return
			}

			members = append(members, Member{s, string(peer.Role()), peer.IsMuted()})
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

		err = httputil.JsonEncode(w, resp)
		if err != nil {
			fmt.Println(err)
		}
	}).Methods("POST")

	r.HandleFunc("/v1/login/start", loginHandlers.Start).Methods("POST")
	r.HandleFunc("/v1/login/pin", loginHandlers.SubmitPin).Methods("POST")
	r.HandleFunc("/v1/login/register", loginHandlers.Register).Methods("POST")

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
