package main

import (
	"local.dev/snippetbox/internal/models"
)

// templateData is a holding structure for any dynamic data we pass to HTML templates.
type templateData struct {
	Snippet models.Snippet
}
