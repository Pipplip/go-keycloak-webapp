package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go-keycloak-webapp/internal/middleware"
	"go-keycloak-webapp/internal/sessions"
)

func TestProfileHandlerEscapesUsernameAndRendersCSRFToken(t *testing.T) {
	store, err := sessions.NewStore(time.Hour)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	sessionID, err := store.Create(sessions.Session{
		UserID:   123,
		Username: "<script>alert(1)</script>",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	session, ok := store.Get(sessionID)
	if !ok {
		t.Fatal("Get() did not return the session")
	}

	profile, err := NewProfileHandler()
	if err != nil {
		t.Fatalf("NewProfileHandler() error = %v", err)
	}
	handler := middleware.RequireAuth(store)(profile)
	request := httptest.NewRequest(http.MethodGet, "/profile", nil)
	request.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	body := response.Body.String()
	if strings.Contains(body, "<script>") {
		t.Fatal("profile rendered unescaped user content")
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Fatal("profile did not HTML-escape the username")
	}
	if !strings.Contains(body, session.CSRFToken) {
		t.Fatal("profile did not render the session CSRF token")
	}
	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want no-store", got)
	}
}
