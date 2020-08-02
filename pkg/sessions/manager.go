package sessions

import (
	"sync"

	"github.com/ephemeral-networks/voicely/pkg/users"
)

// @todo in the future this will be an abstraction over redis

type SessionManager struct {
	sync.Mutex
	sessions map[string]users.User
}

func (sm *SessionManager) NewSession(id string, user users.User) {
	sm.Lock()
	defer sm.Unlock()
	sm.sessions[id] = user
}