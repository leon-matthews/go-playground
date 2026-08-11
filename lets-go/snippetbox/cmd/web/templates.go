package main

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/justinas/nosurf"
	"local.dev/snippetbox/internal/models"
)

// templateData is a holding structure for any dynamic data we pass to HTML templates.
type templateData struct {
	CSRFToken       string
	CurrentYear     int
	Flash           string
	Form            any
	IsAuthenticated bool
	Snippet         models.Snippet
	Snippets        []models.Snippet
}

func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CSRFToken:       nosurf.Token(r),
		CurrentYear:     time.Now().Year(),
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
		IsAuthenticated: app.isAuthenticated(r),
	}
}

// humanDate creates nicely formatted date string
func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

type templateCache map[string]*template.Template

func newTemplateCache() (map[string]*template.Template, error) {
	cache := templateCache{}

	// Find paths to page templates
	pages, err := filepath.Glob("./www/html/pages/*.html")
	if err != nil {
		return nil, err
	}

	// Loop through the page path
	for _, page := range pages {
		// Create empty template with our custom functions
		name := filepath.Base(page)
		ts := template.New(name).Funcs(functions)

		// Parse the base template file into a template set.
		ts.ParseFiles("./www/html/base.html")
		if err != nil {
			return nil, err
		}

		// Add HTML snippets by calling method on the new template set
		ts, err = ts.ParseGlob("./www/html/snippets/*.html")
		if err != nil {
			return nil, err
		}

		// Add page template
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Add the template set to the map as normal...
		cache[name] = ts
	}
	return cache, nil
}
