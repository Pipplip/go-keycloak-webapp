package middleware

import "net/http"

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, ok := CurrentUser(r)
			if !ok {
				http.Error(w, "authentication required", http.StatusUnauthorized)
				return
			}

			for _, assignedRole := range session.Roles {
				if assignedRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}
