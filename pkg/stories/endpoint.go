package stories

import (
	"github.com/gorilla/mux"
)

type Endpoint struct {
}

func NewEndpoint() *Endpoint {
	return &Endpoint{}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	return r
}
