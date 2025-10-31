package mytypes_test

import (
	"testing"

	"mytypes"
)

func TestDouble(t *testing.T) {
	t.Parallel()
	x := mytypes.MyInt(12)
	want := mytypes.MyInt(24)
	p := &x
	p.Double()
	if want != x {
		t.Errorf("got %v, want %v", x, want)
	}
}
