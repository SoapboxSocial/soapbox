package main

import (
	"database/sql"
	"encoding/json"
	"errors"
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
	"github.com/ephemeral-networks/voicely/pkg/api/middleware"
	usersapi "github.com/ephemeral-networks/voicely/pkg/api/users"
	httputil "github.com/ephemeral-networks/voicely/pkg/http"
	"github.com/ephemeral-networks/voicely/pkg/rooms"
	"github.com/ephemeral-networks/voicely/pkg/sessions"
	"github.com/ephemeral-networks/voicely/pkg/users"
)

type RoomPayload struct {
	ID      int            `json:"id"`
	Name    string         `json:"name,omitempty"`
	Members []rooms.Member `json:"members"`
}

type SDPPayload struct {
	Name *string `json:"name,omitempty"`
	ID   *int    `json:"id,omitempty"`
	SDP  string  `json:"sdp"`
	Type string  `json:"type"`
}

type JoinPayload struct {
	Name    string         `json:"name,omitempty"`
	Members []rooms.Member `json:"members"`
	SDP     SDPPayload     `json:"sdp"`
	Role    string         `json:"role"` // @todo find better name
}

func main() {
	db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=voicely password=voicely dbname=voicely sslmode=disable")
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

	usersEndpoints := usersapi.NewUsersEndpoint(ub, s)

	manager := rooms.NewRoomManager()

	r := mux.NewRouter()

	r.HandleFunc("/v1/rooms", func(w http.ResponseWriter, r *http.Request) {
		data := make([]RoomPayload, 0)

		manager.MapRooms(func(room *rooms.Room) {
			if room == nil {
				return
			}

			r := RoomPayload{ID: room.GetID(), Members: make([]rooms.Member, 0)}

			name := room.GetName()
			if name != "" {
				r.Name = name
			}

			room.MapPeers(func(id int, peer rooms.Peer) {
				r.Members = append(r.Members, peer.GetMember())
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

		userID, err := getUserForSession(s, r)
		if err != nil {
			httputil.JsonError(w, 401, httputil.ErrorCodeInvalidRequestBody, "invalid request body")
			return
		}

		user, err := ub.FindByID(userID)
		if err != nil {
			httputil.JsonError(w, 500, httputil.ErrorCodeRoomFailedToJoin, "failed to join room")
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

		sdp, err := room.Join(userID, user.DisplayName, p)
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

		userID, err := getUserForSession(s, r)
		if err != nil {
			httputil.JsonError(w, 401, httputil.ErrorCodeInvalidRequestBody, "invalid request body")
			return
		}

		user, err := ub.FindByID(userID)
		if err != nil {
			httputil.JsonError(w, 500, httputil.ErrorCodeRoomFailedToJoin, "failed to join room")
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

		sdp, err := room.Join(userID, user.DisplayName, p)
		if err != nil {
			httputil.JsonError(w, 500, httputil.ErrorCodeRoomFailedToJoin, "failed to join room")
			return
		}

		members := make([]rooms.Member, 0)

		room.MapPeers(func(id int, peer rooms.Peer) {
			// @todo will need changing
			if id == userID {
				return
			}

			members = append(members, peer.GetMember())
		})

		resp := &JoinPayload{
			Members: members,
			SDP: SDPPayload{
				ID:   &id,
				Type: strings.ToLower(sdp.Type.String()),
				SDP:  sdp.SDP,
			},
			Role: string(room.GetRoleForPeer(userID)),
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

	loginRoutes := r.PathPrefix("/v1/login").Subrouter()
	loginRoutes.HandleFunc("/start", loginHandlers.Start).Methods("POST")
	loginRoutes.HandleFunc("/pin", loginHandlers.SubmitPin).Methods("POST")
	loginRoutes.HandleFunc("/register", loginHandlers.Register).Methods("POST")

	userRoutes := r.PathPrefix("/v1/users").Subrouter()
	userRoutes.HandleFunc("/{id:[0-9]+}", usersEndpoints.GetUserByID).Methods("GET")
	amw := middleware.NewAuthenticationMiddleware(s)
	userRoutes.Use(amw.Middleware)

	log.Fatal(http.ListenAndServe(":8080", r))
}

func getUserForSession(s *sessions.SessionManager, r *http.Request) (int, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return 0, errors.New("no authorization")
	}

	return s.GetUserIDForSession(token)
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
