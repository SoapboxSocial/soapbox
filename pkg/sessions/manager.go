package sessions

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/ephemeral-networks/voicely/pkg/users"
)

type SessionManager struct {
	db *redis.Client
}

func NewSessionManager(db *redis.Client) *SessionManager {
	return &SessionManager{db: db}
}

func (sm *SessionManager) NewSession(id string, user users.User, expiration time.Duration) error {
	return sm.db.Set(context.Background(), generateSessionKey(id), user.ID, expiration).Err() // @todo we may not need more
}

func (sm *SessionManager) GetUserIDForSession(id string) (int, error) {
	str, err := sm.db.Get(context.Background(), generateSessionKey(id)).Result()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(str)
}

func generateSessionKey(id string) string {
	return "session_" + id
}