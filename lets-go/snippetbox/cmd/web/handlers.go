package main

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
)

// application holds dependencies for the web app.
type application struct {
	logger *slog.Logger
}

func (app *application) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	// Create template set.
	files := []string{
		"./www/html/base.html",
		"./www/html/snippets/nav.html",
		"./www/html/pages/index.html",
	}
	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// Write template content as response body, with no data for now
	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Display a form for creating a new snippet..."))
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Save a new snippet..."))
}
