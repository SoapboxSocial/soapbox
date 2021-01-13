package blocks

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Endpoint struct {
}

func NewEndpoint() *Endpoint {
	return &Endpoint{}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/blocks", e.blocks).Methods("GET")
	r.HandleFunc("/blocks/create", e.blockUser).Methods("POST")

	return r
}

func (e *Endpoint) blockUser(w http.ResponseWriter, r *http.Request) {

}

func (e *Endpoint) blocks(w http.ResponseWriter, r *http.Request) {

}
