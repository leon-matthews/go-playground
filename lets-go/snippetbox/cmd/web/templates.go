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
		// Collect all paths needed for this page
		files := []string{
			"./www/html/base.html",
			"./www/html/snippets/nav.html",
			page,
		}

		// Parse the files into a template set.
		ts, err := template.ParseFiles(files...)
		if err != nil {
			return nil, err
		}

		// Save the template set using base name as key, eg. "index.html"
		name := filepath.Base(page)
		cache[name] = ts
	}

	return cache, nil
}
