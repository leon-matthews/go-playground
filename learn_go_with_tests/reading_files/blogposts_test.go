package blogposts_test

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	blogposts "reading_files"
)

func TestNewBlogPosts(t *testing.T) {
	const (
		firstBody = `Title: Post 1
Description: Description 1
Tags: tdd, go
---
Hello blog!`
		secondBody = `Title: Post 2
Description: Description 2
Tags: python, benchmark
---
This is
a story
about benchmarks`
	)

	fs := fstest.MapFS{
		"hello world.md":  {Data: []byte(firstBody)},
		"hello-world2.md": {Data: []byte(secondBody)},
	}

	posts, err := blogposts.NewPostsFromFS(fs)
	if err != nil {
		t.Fatal(err)
	}

	if len(posts) != len(fs) {
		t.Errorf("got %d posts, wanted %d posts", len(posts), len(fs))
	}

	// First
	want := blogposts.Post{
		Title:       "Post 1",
		Description: "Description 1",
		Tags:        []string{"tdd", "go"},
		Body:        "Hello blog!",
	}
	assert.Equal(t, want, posts[0])

	// Second
	want = blogposts.Post{
		Title:       "Post 2",
		Description: "Description 2",
		Tags:        []string{"python", "benchmark"},
		Body:        "This is\na story\nabout benchmarks",
	}
	assert.Equal(t, posts[1], want)
}
