// Package blogposts renders HTML blog posts from folder of plain-text data
package blogposts

import (
	"io/fs"
	"log"
)

// Post represents a single blog post
type Post struct {
	Title       string
	Description string
	Tags        []string
	Body        string
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
