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

// Render writes HTML for the given post
func Render(w io.Writer, post Post) error {
	templ, err := template.ParseFS(postTemplates, "templates/*.html")
	if err != nil {
		return err
	}

	if err := templ.ExecuteTemplate(w, "post.html", post); err != nil {
		return err
	}

	return nil
}
