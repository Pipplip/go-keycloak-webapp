package sessions

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

type Session struct {
	ID        string
	UserID    int64
	Username  string
	Email     string
	Roles     []string
	CSRFToken string
}

type storedSession struct {
	session   Session
	expiresAt time.Time
}

type Store struct {
	mu       sync.Mutex
	lifetime time.Duration
	sessions map[string]storedSession
}

func NewStore(lifetime time.Duration) (*Store, error) {
	if lifetime < time.Second {
		return nil, errors.New("session lifetime must be at least one second")
	}

	return &Store{
		lifetime: lifetime,
		sessions: make(map[string]storedSession),
	}, nil
}

func (s *Store) Create(session Session) (string, error) {
	id, err := generateID()
	if err != nil {
		return "", err
	}
	csrfToken, err := generateID()
	if err != nil {
		return "", err
	}

	session.ID = id
	session.CSRFToken = csrfToken
	session.Roles = append([]string(nil), session.Roles...)

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, value := range s.sessions {
		if !now.Before(value.expiresAt) {
			delete(s.sessions, key)
		}
	}

	s.sessions[id] = storedSession{
		session:   session,
		expiresAt: now.Add(s.lifetime),
	}

	return id, nil
}

func (s *Store) Get(id string) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, ok := s.sessions[id]
	if !ok {
		return Session{}, false
	}
	if !time.Now().Before(stored.expiresAt) {
		delete(s.sessions, id)
		return Session{}, false
	}

	session := stored.session
	session.Roles = append([]string(nil), session.Roles...)
	return session, true
}

func (s *Store) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
}

func (s *Store) Lifetime() time.Duration {
	return s.lifetime
}

func generateID() (string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
