package main

import (
	"log/slog"
	"net/http"
	"strings"
)

// isFromHTMX returns true if the given request was from HTMX
func isFromHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// redirect user to new URL, performing a full reload of that page.
func redirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	// Redirects can be tricky when using HTMX, you have to take care not
	// to just swap in content from the new page.
	if isFromHTMX(r) {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, url, code)
}

// The application struct holds the dependencies needed for our handlers,
// including a htmlRenderer type.
type application struct {
	logger *slog.Logger
	html   *htmlRenderer
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	err := app.html.render(w, 200, nil, "base", "pages/home.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func (app *application) gopher(w http.ResponseWriter, r *http.Request) {
	width := 100
	err := app.html.render(w, http.StatusOK, width, "partial:image:gopher")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

// Define a user type.
// Fields must be exported to allow references in our templates.
type user struct {
	Name     string
	Email    string
	IsGopher bool
}

// Create a hardcoded list of users.
var users = []user{
	{"Alice Madsen", "alice.madsen@example.com", true},
	{"Theo Thatcher", "theo.thatcher@example.com", true},
	{"Maxwell Albright", "maxwell.albright@example.com", false},
	{"Ruby Thompson", "ruby.thompson@example.com", false},
	{"Leona Rowan", "leona.rowan@example.com", false},
	{"Alicia Lennox", "alicia.lennox@example.com", true},
	{"Ruben Mason", "ruben.mason@example.com", false},
	{"Leo Reynolds", "leo.reynolds@example.com", false},
	{"Max Lester", "max.lester@example.com", true},
	{"Theodore Allister", "theodore.allister@example.com", false},
}

func (app *application) listUsers(w http.ResponseWriter, r *http.Request) {
	err := app.html.render(w, 200, users, "base", "pages/users.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}

func (app *application) searchUsers(w http.ResponseWriter, r *http.Request) {
	query := r.FormValue("query")
	query = strings.ToLower(query)

	var matches []user
	if query == "" {
		matches = users
	} else {
		for _, u := range users {
			if strings.Contains(strings.ToLower(u.Name), query) || strings.Contains(strings.ToLower(u.Email), query) {
				matches = append(matches, u)
			}
		}
	}

	// Render the base template by default.
	template := "base"

	// But if the request is coming from HTMX, render the users:rows template instead.
	if isFromHTMX(r) {
		template = "users:rows"
	}

	// Render just the "users:rows" template from the "pages/users.tmpl" file
	// with the matching user details.
	err := app.html.render(w, 200, matches, template, "pages/users.tmpl")
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, http.StatusText(500), 500)
	}
}
