package blogposts_test

import (
	"bytes"
	_ "embed"
	"github.com/stretchr/testify/require"
	"io"
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
	post := blogposts.Post{
		Title:       "hello world",
		Markdown:    "This is a post",
		Description: "This is a description",
		Tags:        []string{"go", "tdd"},
	}

	t.Run("compare rendered post with golden HTML file", func(t *testing.T) {
		buf := bytes.Buffer{}
		want := readFile(t, "post.html")
		renderer, err := blogposts.NewPostRenderer()
		require.NoError(t, err)
		err = renderer.Render(&buf, post)

		if err != nil {
			t.Fatal(err)
		}

		got := buf.String()
		assert.Equal(t, want, got)
	})
}

func BenchmarkRender(b *testing.B) {
	p := blogposts.Post{
		Title:       "hello world",
		Markdown:    "This is a post",
		Description: "This is a description",
		Tags:        []string{"go", "tdd"},
	}
	p.RenderHTML()

	renderinator, err := blogposts.NewPostRenderer()
	require.NoError(b, err)

	b.ResetTimer()
	for b.Loop() {
		renderinator.Render(io.Discard, p)
	}
}
