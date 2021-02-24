package minis

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

	r.HandleFunc("/", e.listMinis).Methods("GET")

	return r
}

func (e *Endpoint) listMinis(http.ResponseWriter, *http.Request) {

}
