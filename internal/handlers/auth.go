package handlers

import (
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"go-keycloak-webapp/internal/auth"
	"go-keycloak-webapp/internal/sessions"
	"go-keycloak-webapp/internal/users"

	"golang.org/x/oauth2"
)

const oidcBrowserCookieName = "oidc_browser"

type keycloakRoleSet struct {
	Roles []string `json:"roles"`
}

type oidcUserClaims struct {
	Email          string                     `json:"email"`
	Username       string                     `json:"preferred_username"`
	RealmAccess    keycloakRoleSet            `json:"realm_access"`
	ResourceAccess map[string]keycloakRoleSet `json:"resource_access"`
}

type AuthHandler struct {
	OIDC          *auth.Provider
	State         *auth.StateStore
	Sessions      *sessions.Store
	Users         users.Repository
	SecureCookies bool
}

func (h *AuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	browserBinding := ""
	if cookie, err := r.Cookie(oidcBrowserCookieName); err == nil &&
		isBrowserBinding(cookie.Value) {
		browserBinding = cookie.Value
	} else {
		var err error
		browserBinding, err = auth.GenerateBrowserBinding()
		if err != nil {
			log.Printf("generate OIDC browser binding: %v", err)
			http.Error(w, "unable to start login", http.StatusInternalServerError)
			return
		}
	}

	transaction, err := h.State.Create(browserBinding)
	if err != nil {
		log.Printf("create OIDC login transaction: %v", err)
		http.Error(w, "unable to start login", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oidcBrowserCookieName,
		Value:    browserBinding,
		Path:     "/",
		Expires:  time.Now().Add(10 * time.Minute),
		MaxAge:   600,
		HttpOnly: true,
		Secure:   h.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})

	url := h.OIDC.OAuth2Config.AuthCodeURL(
		transaction.State,
		oauth2.SetAuthURLParam("nonce", transaction.Nonce),
		oauth2.S256ChallengeOption(transaction.PKCEVerifier),
	)

	http.Redirect(w, r, url, http.StatusFound)
}

func (h *AuthHandler) Callback(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Cache-Control", "no-store")

	state := r.URL.Query().Get("state")
	browserCookie, err := r.Cookie(oidcBrowserCookieName)
	if err != nil {
		http.Error(w, "invalid login transaction", http.StatusBadRequest)
		return
	}

	transaction, ok := h.State.Take(state, browserCookie.Value)
	if !ok {
		http.Error(w, "invalid login transaction", http.StatusBadRequest)
		return
	}

	if r.URL.Query().Get("error") != "" {
		http.Error(w, "login was not completed", http.StatusUnauthorized)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing authorization code", http.StatusBadRequest)
		return
	}

	token, err := h.OIDC.OAuth2Config.Exchange(
		r.Context(),
		code,
		oauth2.VerifierOption(transaction.PKCEVerifier),
	)
	if err != nil {
		log.Printf("exchange OIDC authorization code: %v", err)
		http.Error(w, "login failed", http.StatusBadGateway)
		return
	}

	rawIDToken, ok := token.Extra(
		"id_token",
	).(string)
	if !ok {
		http.Error(w, "login failed", http.StatusBadGateway)
		return
	}

	idToken, err := h.OIDC.Verifier.Verify(
		r.Context(),
		rawIDToken,
	)
	if err != nil {
		log.Printf("verify OIDC ID token: %v", err)
		http.Error(w, "login failed", http.StatusUnauthorized)
		return
	}

	if err := h.OIDC.ValidateIssuer(idToken.Issuer); err != nil {
		log.Printf("validate OIDC ID token issuer: %v", err)
		http.Error(w, "login failed", http.StatusUnauthorized)
		return
	}

	if idToken.Nonce != transaction.Nonce {
		http.Error(w, "invalid login transaction", http.StatusUnauthorized)
		return
	}

	var claims oidcUserClaims
	err = idToken.Claims(&claims)
	if err != nil {
		log.Printf("parse OIDC ID token claims: %v", err)
		http.Error(w, "login failed", http.StatusBadGateway)
		return
	}

	if idToken.Subject == "" {
		http.Error(w, "login response has no subject", http.StatusBadGateway)
		return
	}

	appUser, err := h.Users.UpsertOIDCUser(
		r.Context(),
		users.OIDCIdentity{
			Issuer:   idToken.Issuer,
			Subject:  idToken.Subject,
			Username: claims.Username,
			Email:    claims.Email,
		},
	)
	if err != nil {
		log.Printf("upsert application user: %v", err)
		http.Error(w, "login temporarily unavailable", http.StatusServiceUnavailable)
		return
	}

	sessionID, err := h.Sessions.Create(
		sessions.Session{
			UserID:   appUser.ID,
			Username: claims.Username,
			Email:    claims.Email,
			Roles:    keycloakRolesForClient(claims, h.OIDC.OAuth2Config.ClientID),
		},
	)
	if err != nil {
		log.Printf("create application session: %v", err)
		http.Error(w, "login failed", http.StatusInternalServerError)
		return
	}

	sessionLifetime := h.Sessions.Lifetime()
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  time.Now().Add(sessionLifetime),
		MaxAge:   int(sessionLifetime / time.Second),
		HttpOnly: true,
		Secure:   h.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(
		w,
		cookie,
	)

	http.Redirect(
		w,
		r,
		"/profile",
		http.StatusFound,
	)
}

func isBrowserBinding(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func keycloakRolesForClient(claims oidcUserClaims, clientID string) []string {
	clientRoles := claims.ResourceAccess[clientID].Roles
	roles := make([]string, 0, len(claims.RealmAccess.Roles)+len(clientRoles))
	seen := make(map[string]struct{}, len(claims.RealmAccess.Roles)+len(clientRoles))
	appendRole := func(role string) {
		if role == "" {
			return
		}
		if _, exists := seen[role]; exists {
			return
		}
		seen[role] = struct{}{}
		roles = append(roles, role)
	}
	for _, role := range claims.RealmAccess.Roles {
		if role != "" {
			appendRole("realm:" + role)
		}
	}
	for _, role := range clientRoles {
		if role != "" {
			appendRole("client:" + clientID + ":" + role)
		}
	}
	return roles
}

func (h *AuthHandler) Logout(
	w http.ResponseWriter,
	r *http.Request,
) {
	cookie, err := r.Cookie("session_id")

	if err == nil {
		h.Sessions.Delete(cookie.Value)
	}

	http.SetCookie(
		w,
		&http.Cookie{
			Name:     "session_id",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   h.SecureCookies,
			SameSite: http.SameSiteLaxMode,
		},
	)

	http.Redirect(
		w,
		r,
		"/",
		http.StatusSeeOther,
	)
}
