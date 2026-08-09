package main

import (
	"html/template"
	"path/filepath"

	"local.dev/snippetbox/internal/models"
)

// templateData is a holding structure for any dynamic data we pass to HTML templates.
type templateData struct {
	Snippet  models.Snippet
	Snippets []models.Snippet
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
		name := filepath.Base(page)

		// Parse the base template file into a template set.
		ts, err := template.ParseFiles("./www/html/base.html")
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
