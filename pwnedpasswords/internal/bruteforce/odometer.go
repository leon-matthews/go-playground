package bruteforce

import (
	"fmt"
	"math/bits"
	"sync"
)

// coord hands out contiguous chunks of one length's candidate space to workers.
type coord struct {
	mu    sync.Mutex
	cur   []int // odometer position of the next chunk to hand out
	done  bool
	base  int
	chunk int
}

// reset positions the coordinator at the start of a new length.
func (c *coord) reset(start []int) {
	c.mu.Lock()
	c.cur = append([]int(nil), start...)
	c.done = false
	c.mu.Unlock()
}

// next returns the start of the next chunk, or ok=false once the length is done.
func (c *coord) next() (start []int, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.done {
		return nil, false
	}
	start = append([]int(nil), c.cur...)
	if !addN(c.cur, c.base, c.chunk) {
		c.done = true
	}
	return start, true
}

// frontier returns a conservatively safe resume point: every candidate before
// it is guaranteed processed. At most workers chunks can still be in flight, so
// rewinding the cursor by that much never skips unprocessed candidates. It
// reports ok=false once the length is done, when the cursor no longer names a
// meaningful position.
func (c *coord) frontier(workers int) (indices []int, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.done {
		return nil, false
	}
	fr := append([]int(nil), c.cur...)
	subN(fr, c.base, workers*c.chunk)
	return fr, true
}

// resumeStart returns the length and odometer indices to begin at. Enumeration
// starts exactly at the resume pattern, or at the first one-character candidate.
func resumeStart(alphabet []byte, resume string) (length int, indices []int, err error) {
	if resume == "" {
		return 1, make([]int, 1), nil
	}
	position := make(map[byte]int, len(alphabet))
	for i, b := range alphabet {
		position[b] = i
	}
	indices = make([]int, len(resume))
	for i := 0; i < len(resume); i++ {
		index, ok := position[resume[i]]
		if !ok {
			return 0, nil, fmt.Errorf("resume byte %q is not in the selected alphabet", resume[i])
		}
		indices[i] = index
	}
	return len(resume), indices, nil
}

// pattern renders odometer indices as their candidate string.
func pattern(indices []int, alphabet []byte) string {
	b := make([]byte, len(indices))
	for i, index := range indices {
		b[i] = alphabet[index]
	}
	return string(b)
}

// advance increments the odometer by one, least-significant digit last. It
// returns false on a complete roll-over.
func advance(indices []int, base int) bool {
	for pos := len(indices) - 1; pos >= 0; pos-- {
		indices[pos]++
		if indices[pos] < base {
			return true
		}
		indices[pos] = 0
	}
	return false
}

// addN adds n to the odometer. It returns false if the value rolls past the last
// candidate of this length, in which case indices is left undefined.
func addN(indices []int, base, n int) bool {
	carry := n
	for pos := len(indices) - 1; pos >= 0 && carry > 0; pos-- {
		total := indices[pos] + carry
		indices[pos] = total % base
		carry = total / base
	}
	return carry == 0
}

// subN subtracts n from the odometer, clamping at all-zeros on underflow.
func subN(indices []int, base, n int) {
	for pos := len(indices) - 1; n > 0 && pos >= 0; pos-- {
		digit := indices[pos] - n%base
		n /= base
		if digit < 0 {
			digit += base
			n++ // borrow
		}
		indices[pos] = digit
	}
	if n > 0 {
		for i := range indices {
			indices[i] = 0
		}
	}
}

// chunkForSpace picks the chunk size for a candidate space of the given size so
// that each worker gets about chunksPerWorker chunks, clamped to [minChunk,
// maxChunk]. A small space therefore still splits across workers instead of
// landing on one, while a huge space keeps chunks bounded.
func chunkForSpace(space uint64, workers int) int {
	target := space / (uint64(workers) * chunksPerWorker)
	switch {
	case target < minChunk:
		return minChunk
	case target > maxChunk:
		return maxChunk
	default:
		return int(target)
	}
}

// powSat returns base**exp, saturating at the maximum uint64 rather than
// wrapping, so callers can compare an astronomically large candidate space
// against ordinary bounds without overflow.
func powSat(base, exp uint64) uint64 {
	result := uint64(1)
	for range exp {
		hi, lo := bits.Mul64(result, base)
		if hi != 0 {
			return ^uint64(0)
		}
		result = lo
	}
	return result
}
