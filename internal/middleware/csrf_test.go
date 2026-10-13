package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"go-keycloak-webapp/internal/sessions"
)

func TestRequireCSRFAllowsValidFormToken(t *testing.T) {
	store, err := sessions.NewStore(time.Hour)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	sessionID, err := store.Create(sessions.Session{UserID: 123})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	session, ok := store.Get(sessionID)
	if !ok {
		t.Fatal("Get() did not return the session")
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := RequireAuth(store)(RequireCSRF(next))

	form := url.Values{"csrf_token": {session.CSRFToken}}
	request := httptest.NewRequest(
		http.MethodPost,
		"/logout",
		strings.NewReader(form.Encode()),
	)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("request status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestRequireCSRFRejectsInvalidToken(t *testing.T) {
	store, err := sessions.NewStore(time.Hour)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	sessionID, err := store.Create(sessions.Session{UserID: 123})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	handler := RequireAuth(store)(RequireCSRF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler was called with an invalid CSRF token")
	})))
	request := httptest.NewRequest(http.MethodPost, "/logout", nil)
	request.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	request.Header.Set("X-CSRF-Token", "invalid")
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("request status = %d, want %d", response.Code, http.StatusForbidden)
	}
}
