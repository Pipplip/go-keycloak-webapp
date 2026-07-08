package handlers

import (
	"go-keycloak-webapp/internal/auth"
	"go-keycloak-webapp/internal/sessions"
	"net/http"
)

type AuthHandler struct {
	OIDC     *auth.Provider
	State    *auth.StateStore
	Sessions *sessions.Store
}

func (h *AuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	state := h.State.Create()

	url := h.OIDC.OAuth2Config.AuthCodeURL(
		state,
	)

	http.Redirect(
		w,
		r,
		url,
		http.StatusFound,
	)
}

func (h *AuthHandler) Callback(
	w http.ResponseWriter,
	r *http.Request,
) {

	state := r.URL.Query().Get("state")

	if !h.State.Validate(state) {
		http.Error(
			w,
			"invalid state",
			http.StatusBadRequest,
		)
		return
	}

	code := r.URL.Query().Get("code")

	if code == "" {
		http.Error(
			w,
			"missing code",
			http.StatusBadRequest,
		)
		return
	}

	token, err := h.OIDC.OAuth2Config.Exchange(
		r.Context(),
		code,
	)

	if err != nil {
		http.Error(
			w,
			"token exchange failed",
			http.StatusInternalServerError,
		)
		return
	}

	rawIDToken, ok := token.Extra(
		"id_token",
	).(string)

	if !ok {
		http.Error(
			w,
			"missing id token",
			http.StatusInternalServerError,
		)
		return
	}

	idToken, err := h.OIDC.Verifier.Verify(
		r.Context(),
		rawIDToken,
	)

	if err != nil {
		http.Error(
			w,
			"invalid id token",
			http.StatusUnauthorized,
		)
		return
	}

	var claims struct {
		Subject string   `json:"sub"`
		Email   string   `json:"email"`
		Name    string   `json:"preferred_username"`
		Roles   []string `json:"roles"`
	}

	err = idToken.Claims(&claims)

	if err != nil {
		http.Error(
			w,
			"cannot parse claims",
			http.StatusInternalServerError,
		)
		return
	}

	sessionID := h.Sessions.Create(
		sessions.Session{
			UserID:   claims.Subject,
			Username: claims.Name,
			Email:    claims.Email,
			Roles:    claims.Roles,
		},
	)

	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
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
			SameSite: http.SameSiteLaxMode,
		},
	)

	http.Redirect(
		w,
		r,
		"/",
		http.StatusFound,
	)
}
