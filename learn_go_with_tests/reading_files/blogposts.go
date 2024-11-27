package blogposts

import (
	"io/fs"
	"log"
)

type Post struct {
	Title       string
	Description string
	Tags        []string
	Body        string
}

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
