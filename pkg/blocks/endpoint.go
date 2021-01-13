package blocks

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Endpoint struct {
	backend *Backend
}

func NewEndpoint(backend *Backend) *Endpoint {
	return &Endpoint{
		backend: backend,
	}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", e.blocks).Methods("GET")
	r.HandleFunc("/{id:[0-9]+}", e.blocks).Methods("DELETE")
	r.HandleFunc("/create", e.blockUser).Methods("POST")

	return r
}

func (e *Endpoint) blockUser(w http.ResponseWriter, r *http.Request) {

}

func (e *Endpoint) blocks(w http.ResponseWriter, r *http.Request) {

}
