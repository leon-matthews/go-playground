package search

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearch(t *testing.T) {
	t.Run("advance enumerates a length in lexicographic order", func(t *testing.T) {
		alphabet := []byte("abc")
		indices := []int{0, 0}
		var got []string
		for {
			got = append(got, pattern(indices, alphabet))
			if !advance(indices, len(alphabet)) {
				break
			}
		}
		want := []string{"aa", "ab", "ac", "ba", "bb", "bc", "ca", "cb", "cc"}
		assert.Equal(t, want, got)
	})

	t.Run("addN matches repeated advance", func(t *testing.T) {
		const base = 7
		for n := range 60 {
			stepwise := []int{0, 0, 0}
			rolled := false
			for range n {
				if !advance(stepwise, base) {
					rolled = true
					break
				}
			}
			jump := []int{0, 0, 0}
			ok := addN(jump, base, n)
			if !rolled {
				require.True(t, ok, "n=%d should not roll over", n)
				assert.Equal(t, stepwise, jump, "n=%d", n)
			}
		}
	})

	t.Run("addN reports roll-over past the last candidate", func(t *testing.T) {
		indices := []int{0, 0} // base 3, indices 0..8
		assert.True(t, addN(indices, 3, 8))
		assert.Equal(t, []int{2, 2}, indices)

		assert.False(t, addN([]int{0, 0}, 3, 9))
	})

	t.Run("subN inverts addN and clamps on underflow", func(t *testing.T) {
		const base = 5
		indices := []int{1, 2, 3}
		require.True(t, addN(indices, base, 20))
		subN(indices, base, 20)
		assert.Equal(t, []int{1, 2, 3}, indices)

		clamp := []int{0, 0, 2}
		subN(clamp, base, 100)
		assert.Equal(t, []int{0, 0, 0}, clamp)
	})

	t.Run("resumeStart validates the pattern against the alphabet", func(t *testing.T) {
		alphabet := []byte("abc")

		length, indices, err := resumeStart(alphabet, "cab")
		require.NoError(t, err)
		assert.Equal(t, 3, length)
		assert.Equal(t, []int{2, 0, 1}, indices)

		_, _, err = resumeStart(alphabet, "az")
		assert.Error(t, err)

		length, indices, err = resumeStart(alphabet, "")
		require.NoError(t, err)
		assert.Equal(t, 1, length)
		assert.Equal(t, []int{0}, indices)
	})

	t.Run("parallel chunking covers every candidate exactly once", func(t *testing.T) {
		alphabet := []byte("abcde")
		base := len(alphabet)
		for _, length := range []int{1, 2, 3} {
			co := &coord{base: base, chunk: 7} // small chunk forces many hand-outs
			co.reset(make([]int, length))

			var mu sync.Mutex
			seen := map[string]int{}
			var wg sync.WaitGroup
			for range 4 {
				wg.Go(func() {
					indices := make([]int, length)
					for {
						start, ok := co.next()
						if !ok {
							return
						}
						copy(indices, start)
						for range co.chunk {
							p := pattern(indices, alphabet)
							mu.Lock()
							seen[p]++
							mu.Unlock()
							if !advance(indices, base) {
								break
							}
						}
					}
				})
			}
			wg.Wait()

			expected := 1
			for range length {
				expected *= base
			}
			assert.Len(t, seen, expected, "length %d should yield base^length candidates", length)
			for p, count := range seen {
				assert.Equal(t, 1, count, "candidate %q should be generated exactly once", p)
			}
		}
	})

	t.Run("powSat computes powers and saturates on overflow", func(t *testing.T) {
		assert.Equal(t, uint64(1), powSat(95, 0))
		assert.Equal(t, uint64(95), powSat(95, 1))
		assert.Equal(t, uint64(9025), powSat(95, 2))
		assert.Equal(t, uint64(857375), powSat(95, 3))
		assert.Equal(t, uint64(1)<<63, powSat(2, 63))
		assert.Equal(t, ^uint64(0), powSat(2, 64), "2**64 overflows and saturates")
		assert.Equal(t, ^uint64(0), powSat(95, 10), "95**10 overflows and saturates")
	})

	t.Run("chunkForSpace splits short lengths and clamps", func(t *testing.T) {
		const workers = 32
		// A one-character space is smaller than the floor, so it clamps up.
		assert.Equal(t, minChunk, chunkForSpace(powSat(95, 1), workers))

		// Length three splits into many chunks instead of landing on one worker.
		space := powSat(95, 3)
		chunk := chunkForSpace(space, workers)
		assert.GreaterOrEqual(t, chunk, minChunk)
		assert.LessOrEqual(t, chunk, maxChunk)
		assert.Greater(t, space/uint64(chunk), uint64(workers),
			"length 3 should yield more chunks than workers")

		// A saturated space is bounded by the ceiling.
		assert.Equal(t, maxChunk, chunkForSpace(^uint64(0), workers))
	})
}
