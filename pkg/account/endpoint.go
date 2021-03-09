package account

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	httputil "github.com/soapboxsocial/soapbox/pkg/http"
	"github.com/soapboxsocial/soapbox/pkg/pubsub"
	"github.com/soapboxsocial/soapbox/pkg/sessions"
)

type Endpoint struct {
	backend  *Backend
	queue    *pubsub.Queue
	sessions *sessions.SessionManager
}

func NewEndpoint(backend *Backend, queue *pubsub.Queue, sessions *sessions.SessionManager) *Endpoint {
	return &Endpoint{
		backend:  backend,
		queue:    queue,
		sessions: sessions,
	}
}

func (e *Endpoint) Router() *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", e.delete).Methods("DELETE")

	return r
}

func (e *Endpoint) delete(w http.ResponseWriter, r *http.Request) {
	id, ok := httputil.GetUserIDFromContext(r.Context())
	if !ok {
		httputil.JsonError(w, http.StatusUnauthorized, httputil.ErrorCodeInvalidRequestBody, "invalid id")
		return
	}

	err := e.backend.DeleteAccount(id)
	if err != nil {
		httputil.JsonError(w, http.StatusInternalServerError, httputil.ErrorCodeNotFound, "failed to delete")
		return
	}

	log.Printf("deleted user %d", id)

	err = e.queue.Publish(pubsub.UserTopic, pubsub.NewDeleteUserEvent(id))
	if err != nil {
		log.Printf("failed to write delete event: %v", err)
	}

	err = e.sessions.CloseSession(r.Header.Get("Authorization"))
	if err != nil {
		log.Printf("failed to close session: %v", err)
	}

	httputil.JsonSuccess(w)
}
