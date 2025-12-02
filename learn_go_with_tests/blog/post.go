package blogposts

import (
	"html/template"

	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

// Post represents a single blog post
type Post struct {
	Title       string
	Description string
	Tags        []string
	Markdown    string        // Markdown in Markdown format
	HTML        template.HTML // Known-safe HTML markup
}

// RenderHTML renders the Markdown field into HTML
func (p *Post) RenderHTML() {
	unsafe := markdown.ToHTML([]byte(p.Markdown), nil, nil)
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	p.HTML = template.HTML(html)
}
