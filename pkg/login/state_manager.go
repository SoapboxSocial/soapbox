package login

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const pinExpiration = 15 * time.Minute

// State represents the user login state
type State struct {
	Email string
	Pin   string
}

// StateManager is responsible for handling the login state of a user
type StateManager struct {
	rdb *redis.Client
}

// NewStateManager creates a new state manager
func NewStateManager(rdb *redis.Client) *StateManager {
	return &StateManager{rdb: rdb}
}

// GetState returns the login state for a given token
func (sm *StateManager) GetState(token string) (*State, error) {
	res, err := sm.rdb.Get(sm.rdb.Context(), key(token)).Result()
	if err != nil {
		return nil, err
	}

	state := &State{}
	err = json.Unmarshal([]byte(res), state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// SetPinState sets the login pin state
func (sm *StateManager) SetPinState(token, email, pin string) error {
	state := &State{
		Pin:   pin,
		Email: email,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	_, err = sm.rdb.Set(sm.rdb.Context(), key(token), data, pinExpiration).Result()
	if err != nil {
		return err
	}

	return nil
}

// SetRegistrationState sets the registration state
func (sm *StateManager) SetRegistrationState(token, email string) error {
	state := &State{
		Email: email,
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	_, err = sm.rdb.Set(sm.rdb.Context(), key(token), data, 0).Result()
	if err != nil {
		return err
	}

	return nil
}

// RemoveState removes the state
func (sm *StateManager) RemoveState(token string) {
	sm.rdb.Del(sm.rdb.Context(), key(token))
}

func key(token string) string {
	return fmt.Sprintf("login_state_%s", token)
}
