package minis

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/http/middlewares"
)

type Endpoint struct {
	backend *Backend
	auth    *middlewares.AuthenticationMiddleware
}

func NewEndpoint(backend *Backend, auth *middlewares.AuthenticationMiddleware) *Endpoint {
	return &Endpoint{
		backend: backend,
		auth:    auth,
	}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/scores", e.saveScores).Methods("POST")

	r.Path("/").Methods("GET").Handler(e.auth.Middleware(http.HandlerFunc(e.listMinis)))

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

func (e *Endpoint) saveScores(w http.ResponseWriter, r *http.Request) {

	httputil.JsonSuccess(w)
}
