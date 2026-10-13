package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"go-keycloak-webapp/internal/auth"
	"go-keycloak-webapp/internal/config"
	"go-keycloak-webapp/internal/database"
	"go-keycloak-webapp/internal/handlers"
	"go-keycloak-webapp/internal/sessions"
	"go-keycloak-webapp/internal/users"
	"go-keycloak-webapp/internal/web"
	"go-keycloak-webapp/migrations"
)

func main() {
	// 1. Konfiguration laden
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx := context.Background()
	db, err := database.Open(ctx, cfg.Database.URL)
	if err != nil {
		log.Fatalf("database initialization failed: %v", err)
	}
	defer db.Close()

	migrationCtx, cancelMigrations := context.WithTimeout(ctx, 15*time.Second)
	err = migrations.Apply(migrationCtx, db)
	cancelMigrations()
	if err != nil {
		log.Fatalf("database migrations failed: %v", err)
	}

	sessionStore, err := sessions.NewStore(cfg.Session.Lifetime)
	if err != nil {
		log.Fatalf("session store initialization failed: %v", err)
	}

	// OIDC discovery initializes the provider before request handling starts.
	oidcProvider, err := auth.NewProvider(
		ctx,
		cfg.OIDC.DiscoveryURL,
		cfg.OIDC.Issuer,
		cfg.OIDC.ClientID,
		cfg.OIDC.ClientSecret,
		cfg.OIDC.RedirectURL,
		cfg.OIDC.AuthorizationURL,
	)
	if err != nil {
		log.Fatalf("oidc initialization failed: %v", err)
	}

	stateStore := auth.NewStateStore()

	authHandler := &handlers.AuthHandler{
		OIDC:          oidcProvider,
		State:         stateStore,
		Sessions:      sessionStore,
		Users:         users.NewSQLRepository(db),
		SecureCookies: cfg.Session.CookieSecure,
	}

	router, err := web.NewRouter(web.RouterDependencies{
		AuthHandler:  authHandler,
		Sessions:     sessionStore,
		OIDCClientID: cfg.OIDC.ClientID,
	})
	if err != nil {
		log.Fatalf("router initialization failed: %v", err)
	}

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	log.Printf(
		"server listening on port %s",
		cfg.Server.Port,
	)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
