package metadata

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/users"
)

type Endpoint struct {
	usersBackend *users.UserBackend
}

func NewEndpoint(usersBackend *users.UserBackend) *Endpoint {
	return &Endpoint{
		usersBackend: usersBackend,
	}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/users/{username}", e.user).Methods("GET")
	//r.HandleFunc("/rooms/{id:[0-9]+}", e.room).Methods("GET")

	return r
}

func (e *Endpoint) user(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	username := params["username"]
	user, err := e.usersBackend.GetUserByUsername(username)
	if err != nil {
		httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeNotFound, "not found")
		return
	}

	err = httputil.JsonEncode(w, user)
	if err != nil {
		log.Printf("failed to encode: %v", err)
	}
}

//func (e *Endpoint) room(w http.ResponseWriter, r *http.Request) {
//	params := mux.Vars(r)
//
//	id, err := strconv.Atoi(params["id"])
//	if err != nil {
//		httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeNotFound, "not found")
//		return
//	}
//
//	resp, err := e.roomService.GetRoom(context.Background(), &pb.RoomQuery{Id: int64(id)})
//	if err != nil {
//		httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeNotFound, "not found")
//		return
//	}
//
//	if resp.Room == nil {
//		httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeNotFound, "not found")
//		return
//	}
//
//	room := resp.Room
//
//	if room.Visibility == pb.Visibility_PRIVATE {
//		httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeNotFound, "not found")
//		return
//	}
//
//	err = httputil.JsonEncode(w, room)
//	if err != nil {
//		log.Printf("failed to encode: %v", err)
//	}
//}
