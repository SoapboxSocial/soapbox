package search

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

	r.Path("/search").Methods("GET").HandlerFunc(e.Search)

	return r
}

func (e *Endpoint) Search(w http.ResponseWriter, r *http.Request) {

}