package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-keycloak-webapp/internal/sessions"
)

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name       string
		roles      []string
		wantStatus int
	}{
		{
			name:       "required client role",
			roles:      []string{"realm:user", "client:go-web-app:admin"},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "realm role does not satisfy client role",
			roles:      []string{"realm:admin"},
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "role for another client is denied",
			roles:      []string{"client:another-client:admin"},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store, err := sessions.NewStore(time.Hour)
			if err != nil {
				t.Fatalf("NewStore() error = %v", err)
			}
			sessionID, err := store.Create(sessions.Session{
				UserID: 123,
				Roles:  test.roles,
			})
			if err != nil {
				t.Fatalf("Create() error = %v", err)
			}

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})
			handler := RequireAuth(store)(
				RequireRole("client:go-web-app:admin")(next),
			)
			request := httptest.NewRequest(http.MethodGet, "/admin", nil)
			request.AddCookie(&http.Cookie{Name: "session_id", Value: sessionID})
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("request status = %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}

func TestRequireRoleRejectsUnauthenticatedRequest(t *testing.T) {
	handler := RequireRole("client:go-web-app:admin")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("next handler was called without an authenticated user")
		}),
	)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/admin", nil))

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("request status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}
