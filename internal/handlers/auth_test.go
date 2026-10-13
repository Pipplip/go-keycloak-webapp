package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"go-keycloak-webapp/internal/auth"

	"golang.org/x/oauth2"
)

func TestLoginStartsBoundPKCETransaction(t *testing.T) {
	handler := &AuthHandler{
		OIDC: &auth.Provider{
			OAuth2Config: oauth2.Config{
				ClientID:    "test-client",
				RedirectURL: "http://localhost:8080/auth/callback",
				Endpoint: oauth2.Endpoint{
					AuthURL: "https://issuer.example/authorize",
				},
			},
		},
		State: auth.NewStateStore(),
	}

	request := httptest.NewRequest(http.MethodGet, "/login", nil)
	response := httptest.NewRecorder()
	handler.Login(response, request)

	if response.Code != http.StatusFound {
		t.Fatalf("Login() status = %d, want %d", response.Code, http.StatusFound)
	}

	location, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	query := location.Query()
	state := query.Get("state")
	if state == "" {
		t.Fatal("redirect is missing state")
	}
	if query.Get("nonce") == "" {
		t.Fatal("redirect is missing nonce")
	}

	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Login() set %d cookies, want 1", len(cookies))
	}
	transaction, ok := handler.State.Take(state, cookies[0].Value)
	if !ok {
		t.Fatal("state store is missing the browser-bound login transaction")
	}

	sum := sha256.Sum256([]byte(transaction.PKCEVerifier))
	wantChallenge := base64.RawURLEncoding.EncodeToString(sum[:])
	if got := query.Get("code_challenge"); got != wantChallenge {
		t.Fatalf("code_challenge = %q, want %q", got, wantChallenge)
	}
	if got := query.Get("code_challenge_method"); got != "S256" {
		t.Fatalf("code_challenge_method = %q, want S256", got)
	}

	if cookies[0].Name != oidcBrowserCookieName || !cookies[0].HttpOnly ||
		cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected browser-binding cookie: %#v", cookies[0])
	}
}

func TestKeycloakRolesForClientCombinesOnlyRelevantRoles(t *testing.T) {
	var claims oidcUserClaims
	err := json.Unmarshal([]byte(`{
		"realm_access": {"roles": ["user", "shared", ""]},
		"resource_access": {
			"go-web-app": {"roles": ["editor", "shared"]},
			"another-client": {"roles": ["admin"]}
		}
	}`), &claims)
	if err != nil {
		t.Fatalf("unmarshal Keycloak claims: %v", err)
	}

	got := keycloakRolesForClient(claims, "go-web-app")
	want := []string{
		"realm:user",
		"realm:shared",
		"client:go-web-app:editor",
		"client:go-web-app:shared",
	}
	if len(got) != len(want) {
		t.Fatalf("roles = %v, want %v", got, want)
	}
	for i, role := range want {
		if got[i] != role {
			t.Fatalf("roles = %v, want %v", got, want)
		}
	}
}
