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
