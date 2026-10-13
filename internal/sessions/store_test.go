package sessions

import (
	"testing"
	"time"
)

func TestCreateGeneratesCSRFTokenAndGetReturnsSession(t *testing.T) {
	store, err := NewStore(time.Hour)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	id, err := store.Create(Session{
		UserID: 123,
		Roles:  []string{"reader"},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	session, ok := store.Get(id)
	if !ok {
		t.Fatal("Get() did not return the new session")
	}
	if session.ID != id || session.UserID != 123 {
		t.Fatalf("Get() returned unexpected session: %#v", session)
	}
	if session.CSRFToken == "" {
		t.Fatal("Create() did not generate a CSRF token")
	}
}

func TestGetDeletesExpiredSession(t *testing.T) {
	store, err := NewStore(time.Hour)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	id, err := store.Create(Session{UserID: 123})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	store.mu.Lock()
	stored := store.sessions[id]
	stored.expiresAt = time.Now().Add(-time.Second)
	store.sessions[id] = stored
	store.mu.Unlock()

	if _, ok := store.Get(id); ok {
		t.Fatal("Get() returned an expired session")
	}
	if _, ok := store.sessions[id]; ok {
		t.Fatal("Get() did not remove the expired session")
	}
}

func TestNewStoreRejectsTooShortLifetime(t *testing.T) {
	if _, err := NewStore(time.Millisecond); err == nil {
		t.Fatal("NewStore() accepted a lifetime shorter than one second")
	}
}
