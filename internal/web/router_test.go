package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-keycloak-webapp/internal/handlers"
	"go-keycloak-webapp/internal/sessions"
)

func TestNewRouterProtectsProfileAndAdminRoutes(t *testing.T) {
	store, err := sessions.NewStore(time.Hour)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	router, err := NewRouter(RouterDependencies{
		AuthHandler:  &handlers.AuthHandler{},
		Sessions:     store,
		OIDCClientID: "test-client",
	})
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}

	for _, path := range []string{"/profile", "/admin"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(
				response,
				httptest.NewRequest(http.MethodGet, path, nil),
			)
			if response.Code != http.StatusFound {
				t.Fatalf("request status = %d, want %d", response.Code, http.StatusFound)
			}
			if got := response.Header().Get("Location"); got != "/login" {
				t.Fatalf("redirect location = %q, want /login", got)
			}
		})
	}
}
