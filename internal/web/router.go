package web

import (
	"errors"
	"fmt"
	"net/http"

	"go-keycloak-webapp/internal/handlers"
	"go-keycloak-webapp/internal/middleware"
	"go-keycloak-webapp/internal/sessions"
)

type RouterDependencies struct {
	AuthHandler  *handlers.AuthHandler
	Sessions     *sessions.Store
	OIDCClientID string
}

func NewRouter(deps RouterDependencies) (http.Handler, error) {
	if deps.AuthHandler == nil {
		return nil, errors.New("auth handler is required")
	}
	if deps.Sessions == nil {
		return nil, errors.New("session store is required")
	}
	if deps.OIDCClientID == "" {
		return nil, errors.New("OIDC client ID is required")
	}

	profile, err := handlers.NewProfileHandler()
	if err != nil {
		return nil, fmt.Errorf("create profile handler: %w", err)
	}

	mux := http.NewServeMux()
	mux.Handle(
		"/profile",
		middleware.RequireAuth(deps.Sessions)(profile),
	)

	admin := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Admin area"))
	})
	adminRole := "client:" + deps.OIDCClientID + ":admin"
	mux.Handle(
		"GET /admin",
		middleware.Chain(
			admin,
			middleware.RequireAuth(deps.Sessions),
			middleware.RequireRole(adminRole),
		),
	)

	mux.HandleFunc("GET /login", deps.AuthHandler.Login)
	mux.HandleFunc("GET /auth/callback", deps.AuthHandler.Callback)
	mux.Handle(
		"POST /logout",
		middleware.Chain(
			http.HandlerFunc(deps.AuthHandler.Logout),
			middleware.RequireAuth(deps.Sessions),
			middleware.RequireCSRF,
		),
	)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from Go"))
	})

	return mux, nil
}
