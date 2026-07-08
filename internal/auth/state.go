package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type StateStore struct {
	mu     sync.Mutex
	values map[string]time.Time
}

func NewStateStore() *StateStore {
	return &StateStore{
		values: make(map[string]time.Time),
	}
}

func (s *StateStore) Create() string {
	bytes := make([]byte, 32)

	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	state := hex.EncodeToString(bytes)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.values[state] = time.Now()

	return state
}

func (s *StateStore) Validate(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	created, exists := s.values[state]

	if !exists {
		return false
	}

	delete(s.values, state)

	return time.Since(created) < 10*time.Minute
}
