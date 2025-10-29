package mytypes

import (
	"testing"
)

func TestStringBuilder(t *testing.T) {
	t.Parallel()
	var sb MyBuilder
	sb.WriteString("Hello, ")
	sb.WriteString("world!")
	want := "Hello, world!"
	got := sb.String()
	if want != got {
		t.Errorf("want %s, got %s", want, got)
	}

	wantLen := 13
	gotLen := sb.Len()
	if wantLen != gotLen {
		t.Errorf("%q: want len %d, got %d", sb.String(),
			wantLen, gotLen)
	}
}
