package sessions

import (
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/soapboxsocial/soapbox/pkg/users"
)

type SessionManager struct {
	db *redis.Client
}

func NewSessionManager(db *redis.Client) *SessionManager {
	return &SessionManager{db: db}
}

func (sm *SessionManager) NewSession(id string, user users.User, expiration time.Duration) error {
	return sm.db.Set(sm.db.Context(), generateSessionKey(id), user.ID, expiration).Err()
}

func (sm *SessionManager) GetUserIDForSession(id string) (int, error) {
	str, err := sm.db.Get(sm.db.Context(), generateSessionKey(id)).Result()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(str)
}

func (sm *SessionManager) CloseSession(id string) error {
	_, err := sm.db.Del(sm.db.Context(), generateSessionKey(id)).Result()
	return err
}

func generateSessionKey(id string) string {
	return "session_" + id
}
