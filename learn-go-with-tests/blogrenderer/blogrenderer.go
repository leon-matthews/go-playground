package blogrenderer

import (
	"embed"
	"html/template"
	"io"
)

var (
	//go:embed "templates/*"
	templates embed.FS
)

type Post struct {
	Title       string
	Description string
	Tags        []string
	Body        string
}

type Renderer struct {
	tmpl *template.Template
}

func NewRenderer() (*Renderer, error) {
	tmpl, err := template.ParseFS(templates, "templates/*.tmpl")
	if err != nil {
		return nil, err
	}
	return &Renderer{tmpl: tmpl}, nil
}

func (r *Renderer) Render(w io.Writer, p Post) error {
	return r.tmpl.ExecuteTemplate(w, "blog.html.tmpl", p)
}
