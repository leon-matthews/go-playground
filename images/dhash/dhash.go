// Package dhash computes 128-bit difference hashes of images.
//
// A difference hash encodes the direction of brightness gradients across a
// small grayscale thumbnail of an image, which makes it robust to scaling,
// aspect-ratio changes, and mild colour shifts. It is a Go port of Ben Hoyt's
// Python dhash package: https://github.com/benhoyt/dhash
package dhash

import (
	"encoding/binary"
	"fmt"
	"image"
	"math/bits"
	"strings"

	"golang.org/x/image/draw"
)

// side is the width and height of the sampled difference grid.
// span is the thumbnail dimension needed to produce that grid (side + 1).
const (
	side = 8
	span = side + 1
)

// Hash is a 128-bit difference hash.
//
// Row holds the horizontal-gradient bits in the high 64 bits of the hash and
// Col holds the vertical-gradient bits in the low 64 bits.
type Hash struct {
	Row uint64
	Col uint64
}

// New returns the difference hash of img.
//
// The image is converted to grayscale, downsized to a span*span thumbnail, and
// reduced to two 64-bit gradient hashes.
func New(img image.Image) Hash {
	grays := grays9(img)
	row, col := rowCol(grays[:], side)
	return Hash{Row: row, Col: col}
}

// Distance returns the number of bits that differ between two hashes.
//
// The result ranges from 0 (identical) to 128 (fully opposite).
func (h Hash) Distance(other Hash) int {
	return bits.OnesCount64(h.Row^other.Row) + bits.OnesCount64(h.Col^other.Col)
}

// Bytes returns the hash as 16 big-endian bytes, row hash followed by column hash.
func (h Hash) Bytes() [16]byte {
	var out [16]byte
	binary.BigEndian.PutUint64(out[0:8], h.Row)
	binary.BigEndian.PutUint64(out[8:16], h.Col)
	return out
}

// String returns the hash as 32 hex digits, row hash followed by column hash.
func (h Hash) String() string {
	return fmt.Sprintf("%016x%016x", h.Row, h.Col)
}

// Grays converts img to grayscale and downsizes it to the span*span sample grid,
// returning the pixel intensities in row-major order.
func Grays(img image.Image) [span * span]uint8 {
	return grays9(img)
}

// Matrix renders one 64-bit hash half as a side*side grid of bits.
//
// The zero and one strings are used to render unset and set bits respectively.
func Matrix(half uint64, zero, one string) string {
	return matrix(half, side, zero, one)
}

// FormatGrays renders a sample grid as a span*span matrix of intensity values.
func FormatGrays(grays [span * span]uint8) string {
	return formatGrays(grays[:], side)
}

// grays9 converts img to a grayscale span*span thumbnail and returns its pixels.
func grays9(img image.Image) [span * span]uint8 {
	bounds := img.Bounds()
	full := image.NewGray(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(full, full.Bounds(), img, bounds.Min, draw.Src)

	small := image.NewGray(image.Rect(0, 0, span, span))
	draw.CatmullRom.Scale(small, small.Bounds(), full, full.Bounds(), draw.Src, nil)

	var grays [span * span]uint8
	for y := range span {
		for x := range span {
			grays[y*span+x] = small.Pix[y*small.Stride+x]
		}
	}
	return grays
}

// rowCol computes the row and column gradient hashes from a size+1 wide grid.
func rowCol(grays []uint8, size int) (row, col uint64) {
	width := size + 1
	for y := range size {
		for x := range size {
			offset := y*width + x
			var rowBit, colBit uint64
			if grays[offset] < grays[offset+1] {
				rowBit = 1
			}
			if grays[offset] < grays[offset+width] {
				colBit = 1
			}
			row = row<<1 | rowBit
			col = col<<1 | colBit
		}
	}
	return row, col
}

// matrix renders the low size*size bits of half as a grid, MSB first.
func matrix(half uint64, size int, zero, one string) string {
	total := size * size
	var builder strings.Builder
	for y := range size {
		for x := range size {
			shift := uint(total - 1 - (y*size + x))
			if (half>>shift)&1 == 1 {
				builder.WriteString(one)
			} else {
				builder.WriteString(zero)
			}
		}
		if y < size-1 {
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}

// formatGrays renders the first (size+1)*(size+1) intensities as a right-aligned grid.
func formatGrays(grays []uint8, size int) string {
	width := size + 1
	var builder strings.Builder
	for y := range width {
		for x := range width {
			fmt.Fprintf(&builder, "%4d", grays[y*width+x])
		}
		if y < width-1 {
			builder.WriteByte('\n')
		}
	}
	return builder.String()
}
