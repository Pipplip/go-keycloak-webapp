package handlers

import (
	"html/template"
	"log"
	"net/http"

	"go-keycloak-webapp/internal/middleware"
)

const profilePage = `<!doctype html>
<html lang="en">
<body>
  <p>Hello {{.Username}}</p>
  <form method="post" action="/logout">
    <input type="hidden" name="csrf_token" value="{{.CSRFToken}}">
    <button type="submit">Log out</button>
  </form>
</body>
</html>`

func NewProfileHandler() (http.Handler, error) {
	profileTemplate, err := template.New("profile").Parse(profilePage)
	if err != nil {
		return nil, err
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := middleware.CurrentUser(r)
		if !ok {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := profileTemplate.Execute(w, user); err != nil {
			log.Printf("render profile: %v", err)
		}
	}), nil
}
