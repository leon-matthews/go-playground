package count_test

import (
	"bytes"
	"count"
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"count": count.Main,
	})
}

func TestScript(t *testing.T) {
	t.Parallel()
	testscript.Run(t, testscript.Params{
		Dir: "testdata/script",
	})
}

func TestLinesCountsLinesInInput(t *testing.T) {
	t.Parallel()
	input := bytes.NewBufferString("1\n2\n3")
	c, err := count.NewCounter(
		count.WithInput(input),
		count.WithOutput(os.Stdout),
	)
	if err != nil {
		t.Fatalf("error creating counter: %v", err)
	}
	want := 3
	got := c.Lines()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}

func TestWithInputFromFiles(t *testing.T) {
	t.Run("one file argument", func(t *testing.T) {
		args := []string{"testdata/three_lines.txt"}
		c, err := count.NewCounter(
			count.WithInputFromArgs(args),
		)
		if err != nil {
			t.Fatal(err)
		}
		want := 3
		got := c.Lines()
		if want != got {
			t.Errorf("want %d, got %d", want, got)
		}
	})

	t.Run("zero files ignored", func(t *testing.T) {
		buf := bytes.NewBufferString("1\n2\n3")
		c, err := count.NewCounter(
			count.WithInput(buf),
			count.WithInputFromArgs([]string{}),
		)
		if err != nil {
			t.Fatal(err)
		}
		want := 3
		got := c.Lines()
		if want != got {
			t.Errorf("want %d, got %d", want, got)
		}
	})
}
