package main

import (
	"context"
	"go-keycloak-webapp/internal/sessions"
	"log"
	"net/http"

	"go-keycloak-webapp/internal/auth"
	"go-keycloak-webapp/internal/config"
	"go-keycloak-webapp/internal/handlers"
	"go-keycloak-webapp/internal/middleware"
)

func main() {
	// 1. Konfiguration laden
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	sessionStore := sessions.NewStore()

	// 2. Kontext für Initialisierung
	ctx := context.Background()

	// 3. OIDC Provider initialisieren
	oidcProvider, err := auth.NewProvider(
		ctx,
		cfg.OIDC.Issuer,
		cfg.OIDC.ClientID,
		cfg.OIDC.ClientSecret,
		cfg.OIDC.RedirectURL,
	)
	if err != nil {
		log.Fatalf("oidc initialization failed: %v", err)
	}

	// 4. State Store für OAuth2 state Parameter
	stateStore := auth.NewStateStore()

	// 5. Handler erzeugen
	authHandler := &handlers.AuthHandler{
		OIDC:     oidcProvider,
		State:    stateStore,
		Sessions: sessionStore,
	}

	// 6. HTTP Routing
	mux := http.NewServeMux()

	profile := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			user, _ := middleware.CurrentUser(r)

			w.Write([]byte(
				"Hello " + user.Username,
			))
		},
	)

	mux.Handle(
		"/profile",
		middleware.RequireAuth(
			sessionStore,
		)(profile),
	)

	mux.HandleFunc(
		"/logout",
		authHandler.Logout,
	)

	mux.HandleFunc(
		"/auth/callback",
		authHandler.Callback,
	)

	mux.HandleFunc(
		"/",
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello from Go"))
		},
	)

	// 7. HTTP Server
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: mux,
	}

	log.Printf(
		"server listening on port %s",
		cfg.Server.Port,
	)

	// 8. Server starten
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
