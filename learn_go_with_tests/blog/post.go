package blogposts

import (
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
	"html/template"
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
	unsafe := blackfriday.Run([]byte(p.Markdown))
	html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
	p.HTML = template.HTML(html)
}
