package minis

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
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

func (e *Endpoint) listMinis(w http.ResponseWriter, r *http.Request) {
	minis, err := e.backend.ListMinis()
	if err != nil {
		httputil.JsonError(w, http.StatusNotFound, httputil.ErrorCodeNotFound, "not found")
		return
	}

	err = httputil.JsonEncode(w, minis)
	if err != nil {
		log.Printf("failed to encode: %v", err)
	}
}
