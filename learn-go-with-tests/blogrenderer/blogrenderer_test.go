package blogrenderer_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/nalgeon/be"

	"local.dev/blogrenderer"
)

// You can update your failing snapshots by setting UPDATE_SNAPS, eg.
// UPDATE_SNAPS=true go test ./...
var snapshot = snaps.WithConfig(snaps.Dir("testdata"))

func TestRender(t *testing.T) {

	var (
		post1 = blogrenderer.Post{
			Title:       "Hello, world!",
			Body:        "This is the post's body.",
			Description: "This is its description.",
			Tags:        []string{"go", "tdd"},
		}
	)

	t.Run("coverts single post into HTML", func(t *testing.T) {
		buf := bytes.Buffer{}
		r, err := blogrenderer.NewRenderer()
		be.Err(t, err, nil)

		err = r.Render(&buf, post1)
		be.Err(t, err, nil)

		snapshot.MatchSnapshot(t, buf.String())
	})
}

// before:
// cpu: AMD Ryzen 9 7950X3D 16-Core Processor
// BenchmarkRender-32    	   14079	     86330 ns/op	   28076 B/op	     266 allocs/op
//
// after:
// cpu: AMD Ryzen 9 7950X3D 16-Core Processor
// BenchmarkRender-32    	  237294	      4554 ns/op	     978 B/op	      30 allocs/op
func BenchmarkRender(b *testing.B) {
	p := blogrenderer.Post{
		Title:       "Hello, world!",
		Body:        "This is the post's body.",
		Description: "This is its description.",
		Tags:        []string{"go", "tdd"},
	}
	r, err := blogrenderer.NewRenderer()
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		r.Render(io.Discard, p)
	}
}
