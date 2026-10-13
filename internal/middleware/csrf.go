package middleware

import (
	"crypto/subtle"
	"net/http"
)

func RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}

		session, ok := CurrentUser(r)
		if !ok {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		token := r.Header.Get("X-CSRF-Token")
		if token == "" {
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			if err := r.ParseForm(); err != nil {
				http.Error(w, "invalid form", http.StatusBadRequest)
				return
			}
			token = r.PostForm.Get("csrf_token")
		}

		if len(token) != len(session.CSRFToken) ||
			subtle.ConstantTimeCompare([]byte(token), []byte(session.CSRFToken)) != 1 {
			http.Error(w, "invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
