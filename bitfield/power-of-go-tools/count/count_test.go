package count_test

import (
	"bytes"
	"count"
	"os"
	"testing"
)

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

func TestWithArgs(t *testing.T) {
	t.Run("one file argument", func(t *testing.T) {
		args := []string{"testdata/three_lines.txt"}
		c, err := count.NewCounter(
			count.WithArgs(args),
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

	t.Run("zero files", func(t *testing.T) {
		c, err := count.NewCounter(
			count.WithArgs([]string{}),
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
