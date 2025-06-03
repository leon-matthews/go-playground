package blogposts_test

import (
	"bytes"
	_ "embed"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	blogposts "reading_files"
)

// readFile loads contents of file under testdata folder
func readFile(t testing.TB, filename string) string {
	t.Helper()
	contents, err := os.ReadFile("testdata/" + filename)
	if err != nil {
		t.Fatal(err)
	}
	return string(contents)
}

func TestRenderer(t *testing.T) {
	aPost := blogposts.Post{
		Title:       "hello world",
		Body:        "This is a post",
		Description: "This is a description",
		Tags:        []string{"go", "tdd"},
	}

	t.Run("it converts a single post into HTML", func(t *testing.T) {
		buf := bytes.Buffer{}
		want := readFile(t, "post.html")
		err := blogposts.Render(&buf, aPost)
		if err != nil {
			t.Fatal(err)
		}

		got := buf.String()
		assert.Equal(t, want, got)
	})
}
