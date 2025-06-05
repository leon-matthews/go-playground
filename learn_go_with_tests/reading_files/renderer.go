package blogposts

import (
	"embed"
	"html/template"
	"io"
)

var (
	//go:embed "templates/*"
	postTemplates embed.FS
)

type PostRenderer struct {
	templ *template.Template
}

func NewPostRenderer() (*PostRenderer, error) {
	templ, err := template.ParseFS(postTemplates, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &PostRenderer{templ: templ}, nil
}

// Render writes HTML for the given post
func (r *PostRenderer) Render(w io.Writer, p Post) error {
	if err := r.templ.ExecuteTemplate(w, "post.html", p); err != nil {
		return err
	}
	return nil
}
