package pwned_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	"pwneddb/pwned"
)

func TestHexStrings(t *testing.T) {
	t.Run("length one", func(t *testing.T) {
		h := pwned.HexStrings(1)
		const expectedLength = 16
		s := slices.Collect(h)

		want := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f"}
		assert.Equal(t, len(s), 16)
		assert.Equal(t, want, s)
	})

	t.Run("length two", func(t *testing.T) {
		h := pwned.HexStrings(2)
		const expectedLength = 256
		s := slices.Collect(h)
		assert.Equal(t, len(s), 256)
		assert.Equal(t, s[0], "00")
		assert.Equal(t, s[expectedLength-1], "ff")
	})

	t.Run("length fifteen", func(t *testing.T) {
		const length = 15
		h := pwned.HexStrings(length)
		var first string
		for s := range h {
			first = s
			break
		}
		assert.Equal(t, first, "000000000000000")
		assert.Equal(t, len(first), length)
	})
}

func TestHexStringsErrors(t *testing.T) {

}
