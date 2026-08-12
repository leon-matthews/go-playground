package main

import (
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"time"

	"github.com/justinas/nosurf"

	"local.dev/snippetbox/internal/models"
	assets "local.dev/snippetbox/www"
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
	// Find paths to page templates
	pages, err := fs.Glob(assets.HTMLFiles, "pages/*.html")
	if err != nil {
		return nil, err
	}

	// Loop through the page path
	cache := templateCache{}
	for _, page := range pages {
		// Create empty template with our custom functions
		name := filepath.Base(page)
		patterns := []string{
			"base.html",
			"snippets/*.html",
			page,
		}
		ts, err := template.New(name).Funcs(functions).ParseFS(assets.HTMLFiles, patterns...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}
