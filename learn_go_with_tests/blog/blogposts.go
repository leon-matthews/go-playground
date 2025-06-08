// Package blogposts renders HTML blog posts from folder of plain-text data
package blogposts

import (
	"html/template"
	"io/fs"
	"log"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
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

// NewPostsFromFS reads all posts from the root of the given folder
func NewPostsFromFS(fileSystem fs.FS) ([]Post, error) {
	dir, _ := fs.ReadDir(fileSystem, ".")
	var posts []Post
	for _, f := range dir {
		post, err := getPost(fileSystem, f.Name())
		if err != nil {
			log.Printf("error getting post %s: %v", f.Name(), err)
			continue
		}
		posts = append(posts, post)
	}
	return posts, nil
}
