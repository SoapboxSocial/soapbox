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

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
	"github.com/sendgrid/sendgrid-go"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pion/webrtc/v3"

	devicesapi "github.com/ephemeral-networks/voicely/pkg/api/devices"
	"github.com/ephemeral-networks/voicely/pkg/api/login"
	"github.com/ephemeral-networks/voicely/pkg/api/middleware"
	usersapi "github.com/ephemeral-networks/voicely/pkg/api/users"
	"github.com/ephemeral-networks/voicely/pkg/devices"
	"github.com/ephemeral-networks/voicely/pkg/followers"
	httputil "github.com/ephemeral-networks/voicely/pkg/http"
	"github.com/ephemeral-networks/voicely/pkg/images"
	"github.com/ephemeral-networks/voicely/pkg/mail"
	"github.com/ephemeral-networks/voicely/pkg/notifications"
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

// @todo do this in config
const sendgrid_api = "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"

// @TODO: THINK ABOUT CHANGING QUEUES TO REDIS PUBSUB

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

	queue := notifications.NewNotificationQueue(rdb)

	s := sessions.NewSessionManager(rdb)
	ub := users.NewUserBackend(db)
	fb := followers.NewFollowersBackend(db)

	client, err := elasticsearch.NewDefaultClient()
	if err != nil {
		panic(err)
	}

	search := users.NewSearchBackend(client)

	devicesBackend := devices.NewDevicesBackend(db)

	manager := rooms.NewRoomManager()

	amw := middleware.NewAuthenticationMiddleware(s)

	r := mux.NewRouter()

	r.MethodNotAllowedHandler = http.HandlerFunc(httputil.NotAllowedHandler)
	r.NotFoundHandler = http.HandlerFunc(httputil.NotFoundHandler)

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

	roomRoutes := r.PathPrefix("/v1/rooms").Methods("POST").Subrouter()

	roomRoutes.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid request body")
			return
		}

		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
			return
		}

		user, err := ub.FindByID(userID)
		if err != nil {
			httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeRoomFailedToJoin, "failed to join room")
			return
		}

		payload := &SDPPayload{}
		err = json.Unmarshal(b, payload)
		if err != nil {
			httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid request body")
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

		sdp, err := room.Join(userID, user.DisplayName, user.Image, p)
		if err != nil {
			manager.RemoveRoom(room.GetID())
			httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeFailedToCreateRoom, "failed to create room")
			return
		}

		id := room.GetID()
		resp := &SDPPayload{ID: &id, Type: strings.ToLower(sdp.Type.String()), SDP: sdp.SDP}

		err = httputil.JsonEncode(w, resp)
		if err != nil {
			manager.RemoveRoom(room.GetID())
			fmt.Println(err)
			return
		}

		queue.Push(notifications.Event{
			Type:    notifications.EventTypeRoomCreation,
			Creator: userID,
			Params:  map[string]interface{}{"name": name, "id": id},
		})

	})

	roomRoutes.HandleFunc("/{id:[0-9]+}/join", func(w http.ResponseWriter, r *http.Request) {
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid request body")
			return
		}

		userID, ok := middleware.GetUserIDFromContext(r.Context())
		if !ok {
			httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "invalid id")
			return
		}

		user, err := ub.FindByID(userID)
		if err != nil {
			httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeRoomFailedToJoin, "failed to join room")
			return
		}

		payload := &SDPPayload{}
		err = json.Unmarshal(b, payload)
		if err != nil {
			httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "invalid request body")
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
			httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeRoomNotFound, "room not found")
			return
		}

		sdp, err := room.Join(userID, user.DisplayName, user.Image, p)
		if err != nil {
			httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeRoomFailedToJoin, "failed to join room")
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
	})
	roomRoutes.Use(amw.Middleware)

	loginRoutes := r.PathPrefix("/v1/login").Methods("POST").Subrouter()

	ib := images.NewImagesBackend("/cdn/images")
	ms := mail.NewMailService(sendgrid.NewSendClient(sendgrid_api))
	loginHandlers := login.NewLoginEndpoint(ub, s, ms, ib)
	loginRoutes.HandleFunc("/start", loginHandlers.Start)
	loginRoutes.HandleFunc("/pin", loginHandlers.SubmitPin)
	loginRoutes.HandleFunc("/register", loginHandlers.Register)

	usersEndpoints := usersapi.NewUsersEndpoint(ub, fb, s, queue, ib, search)

	r.HandleFunc("/v1/users/search", usersEndpoints.Search)

	userRoutes := r.PathPrefix("/v1/users").Subrouter()

	userRoutes.HandleFunc("/{id:[0-9]+}", usersEndpoints.GetUserByID).Methods("GET")
	userRoutes.HandleFunc("/{id:[0-9]+}/followers", usersEndpoints.GetFollowersForUser).Methods("GET")
	userRoutes.HandleFunc("/{id:[0-9]+}/following", usersEndpoints.GetFollowedByForUser).Methods("GET")
	userRoutes.HandleFunc("/follow", usersEndpoints.FollowUser).Methods("POST")
	userRoutes.HandleFunc("/unfollow", usersEndpoints.UnfollowUser).Methods("POST")
	userRoutes.HandleFunc("/edit", usersEndpoints.EditUser).Methods("POST")

	userRoutes.Use(amw.Middleware)

	devicesRoutes := r.PathPrefix("/v1/devices").Subrouter()

	devicesEndpoint := devicesapi.NewDevicesEndpoint(devicesBackend)
	devicesRoutes.HandleFunc("/add", devicesEndpoint.AddDevice).Methods("POST")
	devicesRoutes.Use(amw.Middleware)

	headersOk := handlers.AllowedHeaders([]string{
		"Content-Type",
		"X-Requested-With",
		"Accept",
		"Accept-Language",
		"Accept-Encoding",
		"Content-Language",
		"Origin",
	})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk)(r)))
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
