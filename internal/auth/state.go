package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

const loginTransactionLifetime = 10 * time.Minute

var errMissingBrowserBinding = errors.New("browser binding is required")

type LoginTransaction struct {
	State        string
	Nonce        string
	PKCEVerifier string
}

type storedTransaction struct {
	transaction    LoginTransaction
	browserBinding string
	createdAt      time.Time
}

type StateStore struct {
	mu     sync.Mutex
	values map[string]storedTransaction
}

func NewStateStore() *StateStore {
	return &StateStore{
		values: make(map[string]storedTransaction),
	}
}

func (s *StateStore) Create(browserBinding string) (LoginTransaction, error) {
	if browserBinding == "" {
		return LoginTransaction{}, errMissingBrowserBinding
	}

	state, err := randomValue()
	if err != nil {
		return LoginTransaction{}, err
	}

	nonce, err := randomValue()
	if err != nil {
		return LoginTransaction{}, err
	}

	verifier, err := randomValue()
	if err != nil {
		return LoginTransaction{}, err
	}

	transaction := LoginTransaction{
		State:        state,
		Nonce:        nonce,
		PKCEVerifier: verifier,
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, value := range s.values {
		if now.Sub(value.createdAt) >= loginTransactionLifetime {
			delete(s.values, key)
		}
	}

	s.values[state] = storedTransaction{
		transaction:    transaction,
		browserBinding: browserBinding,
		createdAt:      now,
	}

	return transaction, nil
}

func (s *StateStore) Take(state, browserBinding string) (LoginTransaction, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	stored, exists := s.values[state]
	if !exists {
		return LoginTransaction{}, false
	}

	if time.Since(stored.createdAt) >= loginTransactionLifetime {
		delete(s.values, state)
		return LoginTransaction{}, false
	}

	if len(browserBinding) != len(stored.browserBinding) ||
		subtle.ConstantTimeCompare([]byte(browserBinding), []byte(stored.browserBinding)) != 1 {
		return LoginTransaction{}, false
	}

	delete(s.values, state)
	return stored.transaction, true
}

func GenerateBrowserBinding() (string, error) {
	return randomValue()
}

func randomValue() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
