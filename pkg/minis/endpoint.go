package minis

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/http/middlewares"
)

type Endpoint struct {
	backend *Backend
	auth    *middlewares.AuthenticationMiddleware
	keys    AuthKeys
}

func NewEndpoint(backend *Backend, auth *middlewares.AuthenticationMiddleware, keys AuthKeys) *Endpoint {
	return &Endpoint{
		backend: backend,
		auth:    auth,
		keys:    keys,
	}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/scores", e.saveScores).Methods("POST")

	r.Path("/").Methods("GET").Handler(e.auth.Middleware(http.HandlerFunc(e.listMinis)))

	return r
}

func (e *Endpoint) listMinis(w http.ResponseWriter, _ *http.Request) {
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
	query := r.URL.Query()
	token := query.Get("token")
	id, ok := e.keys[token]
	if !ok {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeUnauthorized, "unauthorized")
		return
	}

	room := query.Get("room")
	if room == "" {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "bad request")
		return
	}

	var scores Scores
	err := json.NewDecoder(r.Body).Decode(&scores)
	if err != nil {
		httputil.JsonError(w, http.StatusBadRequest, httputil.ErrorCodeInvalidRequestBody, "bad request")
		return
	}

	err = e.backend.SaveScores(id, room, scores)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeInvalidRequestBody, "failed")
		return
	}

	httputil.JsonSuccess(w)
}
