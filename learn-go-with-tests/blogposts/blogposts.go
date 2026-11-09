package blogposts

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"strings"
)

const (
	titleSeparator       = "Title: "
	descriptionSeparator = "Description: "
	tagSeparator         = "Tags: "
)

type Post struct {
	Title       string
	Description string
	Tags        []string
	Body        string
}

// NewPostsFromFS loads posts from the current directory
func NewPostsFromFS(fsys fs.FS) ([]Post, error) {
	dirEntries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("finding posts: %w", err)
	}

	posts := make([]Post, 0, len(dirEntries))
	for _, d := range dirEntries {
		post, err := getPost(fsys, d.Name())
		if err != nil {
			return nil, fmt.Errorf("reading post: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// getPost opens file and reads post
func getPost(fsys fs.FS, path string) (Post, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return Post{}, err
	}
	defer f.Close()
	return readPost(f)
}

// readPost creates post from reader
func readPost(r io.Reader) (Post, error) {
	scanner := bufio.NewScanner(r)

	// Medatadata
	readMeta := func(separator string) string {
		scanner.Scan()
		return strings.TrimPrefix(scanner.Text(), separator)
	}

	post := Post{
		Title:       readMeta(titleSeparator),
		Description: readMeta(descriptionSeparator),
		Tags:        strings.Split(readMeta(tagSeparator), ", "),
		Body:        readBody(scanner),
	}
	return post, nil
}

// readBody creates body from scanner
func readBody(scanner *bufio.Scanner) string {
	scanner.Scan() // Skip "---"
	b := strings.Builder{}
	for scanner.Scan() {
		fmt.Fprintln(&b, scanner.Text())
	}
	return strings.TrimSuffix(b.String(), "\n")
}
