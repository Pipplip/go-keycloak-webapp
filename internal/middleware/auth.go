package middleware

import (
	"context"
	"net/http"

	"go-keycloak-webapp/internal/sessions"
)

type contextKey string

const userContextKey contextKey = "user"

func RequireAuth(
	store *sessions.Store,
) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(
			w http.ResponseWriter,
			r *http.Request,
		) {

			cookie, err := r.Cookie("session_id")

			if err != nil {
				http.Redirect(
					w,
					r,
					"/login",
					http.StatusFound,
				)
				return
			}

			session, exists := store.Get(
				cookie.Value,
			)

			if !exists {
				http.Redirect(
					w,
					r,
					"/login",
					http.StatusFound,
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				userContextKey,
				session,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)

		})
	}
}

func CurrentUser(
	r *http.Request,
) (sessions.Session, bool) {

	user, ok := r.Context().
		Value(userContextKey).(sessions.Session)

	return user, ok
}
