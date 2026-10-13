package config

import (
	"testing"
	"time"
)

func TestLoadSessionSettings(t *testing.T) {
	setRequiredOIDCEnvironment(t)
	t.Setenv("APP_SESSION_TTL", "12h")
	t.Setenv("APP_SESSION_COOKIE_SECURE", "true")
	t.Setenv("OIDC_AUTHORIZATION_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Session.Lifetime != 12*time.Hour {
		t.Fatalf("session lifetime = %s, want 12h", cfg.Session.Lifetime)
	}
	if !cfg.Session.CookieSecure {
		t.Fatal("CookieSecure = false, want true")
	}
	if cfg.OIDC.DiscoveryURL != cfg.OIDC.Issuer {
		t.Fatalf("default discovery URL = %q, want issuer %q", cfg.OIDC.DiscoveryURL, cfg.OIDC.Issuer)
	}
}

func TestLoadSeparatesPublicIssuerFromDiscoveryURL(t *testing.T) {
	setRequiredOIDCEnvironment(t)
	t.Setenv("OIDC_ISSUER", "http://localhost:8081/realms/test")
	t.Setenv("OIDC_DISCOVERY_URL", "http://keycloak:8080/realms/test")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.OIDC.Issuer != "http://localhost:8081/realms/test" {
		t.Fatalf("issuer = %q", cfg.OIDC.Issuer)
	}
	if cfg.OIDC.DiscoveryURL != "http://keycloak:8080/realms/test" {
		t.Fatalf("discovery URL = %q", cfg.OIDC.DiscoveryURL)
	}
}

func TestLoadRejectsInvalidSessionSettings(t *testing.T) {
	setRequiredOIDCEnvironment(t)
	t.Setenv("APP_SESSION_TTL", "not-a-duration")
	t.Setenv("APP_SESSION_COOKIE_SECURE", "false")
	t.Setenv("OIDC_AUTHORIZATION_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted an invalid session lifetime")
	}

	t.Setenv("APP_SESSION_TTL", "8h")
	t.Setenv("APP_SESSION_COOKIE_SECURE", "not-a-bool")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted an invalid secure-cookie setting")
	}

	t.Setenv("APP_SESSION_COOKIE_SECURE", "false")
	t.Setenv("OIDC_AUTHORIZATION_URL", "keycloak/auth")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a relative authorization URL")
	}

	t.Setenv("OIDC_AUTHORIZATION_URL", "")
	t.Setenv("OIDC_DISCOVERY_URL", "relative/discovery")
	if _, err := Load(); err == nil {
		t.Fatal("Load() accepted a relative discovery URL")
	}
}

func setRequiredOIDCEnvironment(t *testing.T) {
	t.Helper()
	t.Setenv("OIDC_ISSUER", "https://issuer.example/realms/test")
	t.Setenv("OIDC_CLIENT_ID", "test-client")
	t.Setenv("OIDC_CLIENT_SECRET", "test-secret")
	t.Setenv("OIDC_REDIRECT_URL", "http://localhost:8080/auth/callback")
	t.Setenv("APP_DATABASE_URL", "postgres://app:secret@localhost:5432/app?sslmode=disable")
}
