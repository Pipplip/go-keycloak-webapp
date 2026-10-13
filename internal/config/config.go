package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	OIDC     OIDCConfig
	Session  SessionConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port string
}

type OIDCConfig struct {
	Issuer           string
	DiscoveryURL     string
	ClientID         string
	ClientSecret     string
	RedirectURL      string
	AuthorizationURL string
}

type SessionConfig struct {
	Lifetime     time.Duration
	CookieSecure bool
}

type DatabaseConfig struct {
	URL string
}

func Load() (*Config, error) {
	issuer := os.Getenv("OIDC_ISSUER")
	discoveryURL := getEnv("OIDC_DISCOVERY_URL", issuer)

	sessionLifetime, err := time.ParseDuration(
		getEnv("APP_SESSION_TTL", "8h"),
	)
	if err != nil {
		return nil, fmt.Errorf("APP_SESSION_TTL: %w", err)
	}

	cookieSecure, err := strconv.ParseBool(
		getEnv("APP_SESSION_COOKIE_SECURE", "false"),
	)
	if err != nil {
		return nil, fmt.Errorf("APP_SESSION_COOKIE_SECURE: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("APP_PORT", "8080"),
		},
		OIDC: OIDCConfig{
			Issuer:           issuer,
			DiscoveryURL:     discoveryURL,
			ClientID:         os.Getenv("OIDC_CLIENT_ID"),
			ClientSecret:     os.Getenv("OIDC_CLIENT_SECRET"),
			RedirectURL:      os.Getenv("OIDC_REDIRECT_URL"),
			AuthorizationURL: os.Getenv("OIDC_AUTHORIZATION_URL"),
		},
		Session: SessionConfig{
			Lifetime:     sessionLifetime,
			CookieSecure: cookieSecure,
		},
		Database: DatabaseConfig{
			URL: os.Getenv("APP_DATABASE_URL"),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.OIDC.Issuer == "" {
		return fmt.Errorf("OIDC_ISSUER is required")
	}
	if !validHTTPURL(c.OIDC.Issuer) {
		return fmt.Errorf("OIDC_ISSUER must be an absolute HTTP or HTTPS URL")
	}
	if !validHTTPURL(c.OIDC.DiscoveryURL) {
		return fmt.Errorf("OIDC_DISCOVERY_URL must be an absolute HTTP or HTTPS URL")
	}

	if c.OIDC.ClientID == "" {
		return fmt.Errorf("OIDC_CLIENT_ID is required")
	}

	if c.OIDC.ClientSecret == "" {
		return fmt.Errorf("OIDC_CLIENT_SECRET is required")
	}

	if c.OIDC.RedirectURL == "" {
		return fmt.Errorf("OIDC_REDIRECT_URL is required")
	}

	if c.OIDC.AuthorizationURL != "" {
		authorizationURL, err := url.Parse(c.OIDC.AuthorizationURL)
		if err != nil || !authorizationURL.IsAbs() ||
			(authorizationURL.Scheme != "http" && authorizationURL.Scheme != "https") ||
			authorizationURL.Host == "" {
			return fmt.Errorf("OIDC_AUTHORIZATION_URL must be an absolute HTTP or HTTPS URL")
		}
	}

	if c.Session.Lifetime < time.Second {
		return fmt.Errorf("APP_SESSION_TTL must be at least one second")
	}

	if c.Database.URL == "" {
		return fmt.Errorf("APP_DATABASE_URL is required")
	}

	return nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	return value
}

func validHTTPURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil &&
		parsed.IsAbs() &&
		(parsed.Scheme == "http" || parsed.Scheme == "https") &&
		parsed.Host != ""
}
