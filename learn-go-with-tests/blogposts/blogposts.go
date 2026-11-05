package blogposts

import (
	"fmt"
	"io/fs"
)

type Post struct {
	Title string
}

func NewPostsFromFS(fileSystem fs.FS) ([]Post, error) {
	dirEntries, err := fs.ReadDir(fileSystem, ".")
	if err != nil {
		return nil, fmt.Errorf("finding posts: %w", err)
	}

	posts := make([]Post, 0, len(dirEntries))
	for _, d := range dirEntries {
		title := d.Name()
		posts = append(posts, Post{Title: title})
	}
	return posts, nil
}
