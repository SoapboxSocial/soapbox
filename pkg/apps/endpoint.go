package apps

import "github.com/gorilla/mux"

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

	//r.HandleFunc("/", e.unblock).Methods("DELETE")
	//r.HandleFunc("/create", e.block).Methods("POST")

	return r
}
