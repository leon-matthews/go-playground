package dhash

import (
	"image"
	"strings"
	"testing"
)

// doctestImage is the 5x5 sample grid from Ben Hoyt's dhash doctests (size=4).
var doctestImage = []uint8{
	0, 0, 1, 1, 1,
	0, 1, 1, 3, 4,
	0, 1, 6, 6, 7,
	7, 7, 7, 7, 9,
	8, 7, 7, 8, 9,
}

func TestRowCol(t *testing.T) {
	row, col := rowCol(doctestImage, 4)
	if row != 0x4bd1 {
		t.Errorf("row = %#04x, want %#04x", row, 0x4bd1)
	}
	if col != 0x53f9 {
		t.Errorf("col = %#04x, want %#04x", col, 0x53f9)
	}
}

func TestDistance(t *testing.T) {
	tests := []struct {
		a, b Hash
		want int
	}{
		{Hash{Row: 0x4bd1}, Hash{Row: 0x4bd1}, 0},
		{Hash{Row: 0x4bd1}, Hash{Row: 0x5bd2}, 3},
		{Hash{}, Hash{Row: ^uint64(0), Col: ^uint64(0)}, 128},
	}
	for _, test := range tests {
		if got := test.a.Distance(test.b); got != test.want {
			t.Errorf("Distance(%v, %v) = %d, want %d", test.a, test.b, got, test.want)
		}
	}
}

func TestString(t *testing.T) {
	h := Hash{Row: 0x4bd1, Col: 0x3a6f}
	want := strings.Repeat("0", 12) + "4bd1" + strings.Repeat("0", 12) + "3a6f"
	if got := h.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestBytes(t *testing.T) {
	h := Hash{Row: 0x4bd1, Col: 0x3a6f}
	want := [16]byte{0, 0, 0, 0, 0, 0, 0x4b, 0xd1, 0, 0, 0, 0, 0, 0, 0x3a, 0x6f}
	if got := h.Bytes(); got != want {
		t.Errorf("Bytes() = %v, want %v", got, want)
	}
}

func TestMatrix(t *testing.T) {
	row, col := rowCol(doctestImage, 4)

	wantRow := ".*..\n*.**\n**.*\n...*"
	if got := matrix(row, 4, ".", "*"); got != wantRow {
		t.Errorf("row matrix =\n%s\nwant\n%s", got, wantRow)
	}

	wantCol := "0101\n0011\n1111\n1001"
	if got := matrix(col, 4, "0", "1"); got != wantCol {
		t.Errorf("col matrix =\n%s\nwant\n%s", got, wantCol)
	}
}

func TestFormatGrays(t *testing.T) {
	want := strings.Join([]string{
		"   0   0   1   1   1",
		"   0   1   1   3   4",
		"   0   1   6   6   7",
		"   7   7   7   7   9",
		"   8   7   7   8   9",
	}, "\n")
	if got := formatGrays(doctestImage, 4); got != want {
		t.Errorf("formatGrays =\n%s\nwant\n%s", got, want)
	}
}

// TestNewGradient checks the full pipeline: a horizontal ramp must set every row
// bit and clear every column bit, and a vertical ramp must do the reverse.
func TestNewGradient(t *testing.T) {
	horizontal := image.NewGray(image.Rect(0, 0, 180, 90))
	for y := range 90 {
		for x := range 180 {
			horizontal.Pix[y*horizontal.Stride+x] = uint8(x * 255 / 179)
		}
	}
	if h := New(horizontal); h.Row != ^uint64(0) || h.Col != 0 {
		t.Errorf("horizontal ramp: row=%#016x col=%#016x, want row=all-ones col=0", h.Row, h.Col)
	}

	vertical := image.NewGray(image.Rect(0, 0, 90, 180))
	for y := range 180 {
		for x := range 90 {
			vertical.Pix[y*vertical.Stride+x] = uint8(y * 255 / 179)
		}
	}
	if h := New(vertical); h.Col != ^uint64(0) || h.Row != 0 {
		t.Errorf("vertical ramp: row=%#016x col=%#016x, want row=0 col=all-ones", h.Row, h.Col)
	}
}
