package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewProviderSeparatesPublicIssuerFromInternalEndpoints(t *testing.T) {
	var discoveryURL string
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		discoveryURL = r.URL.Path
		if err := json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                server.URL,
			"authorization_endpoint":                server.URL + "/authorize",
			"token_endpoint":                        server.URL + "/token",
			"jwks_uri":                              server.URL + "/keys",
			"response_types_supported":              []string{"code"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}); err != nil {
			t.Errorf("encode discovery document: %v", err)
		}
	}))
	defer server.Close()

	const publicIssuer = "http://localhost:8081/realms/test"
	const browserAuthorizationURL = "http://localhost:8081/realms/test/protocol/openid-connect/auth"
	provider, err := NewProvider(
		context.Background(),
		server.URL,
		publicIssuer,
		"test-client",
		"test-secret",
		"http://localhost:8080/auth/callback",
		browserAuthorizationURL,
	)
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}

	if discoveryURL != "/.well-known/openid-configuration" {
		t.Fatalf("discovery path = %q, want /.well-known/openid-configuration", discoveryURL)
	}
	if got := provider.OAuth2Config.Endpoint.AuthURL; got != browserAuthorizationURL {
		t.Fatalf("authorization URL = %q, want %q", got, browserAuthorizationURL)
	}
	if got, want := provider.OAuth2Config.Endpoint.TokenURL, server.URL+"/token"; got != want {
		t.Fatalf("token URL = %q, want %q", got, want)
	}
	if provider.Issuer != publicIssuer {
		t.Fatalf("provider issuer = %q, want %q", provider.Issuer, publicIssuer)
	}
	if err := provider.ValidateIssuer(publicIssuer); err != nil {
		t.Fatalf("ValidateIssuer() rejected the configured issuer: %v", err)
	}
	if err := provider.ValidateIssuer(server.URL); err == nil {
		t.Fatal("ValidateIssuer() accepted the discovery host as token issuer")
	}
}
