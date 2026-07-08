package sessions

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type Session struct {
	ID       string
	UserID   string
	Username string
	Email    string
	Roles    []string
}

type Store struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[string]Session),
	}
}

func (s *Store) Create(session Session) string {
	id := generateID()

	session.ID = id

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[id] = session

	return id
}

func (s *Store) Get(id string) (Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[id]

	return session, ok
}

func (s *Store) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
}

func generateID() string {
	bytes := make([]byte, 32)

	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(bytes)
}
