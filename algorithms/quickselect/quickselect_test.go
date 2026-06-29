package quickselect

import (
	"cmp"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNthElement(t *testing.T) {
	t.Parallel()
	unordered := []int{5, 3, 1, 4, 2}

	t.Run("returns the minimum at k=0", func(t *testing.T) {
		assert.Equal(t, 1, NthElement(slices.Clone(unordered), 0))
	})

	t.Run("returns the maximum at k=len-1", func(t *testing.T) {
		assert.Equal(t, 5, NthElement(slices.Clone(unordered), len(unordered)-1))
	})

	t.Run("returns the kth smallest for every k", func(t *testing.T) {
		for k := range len(unordered) {
			values := slices.Clone(unordered)
			assert.Equal(t, k+1, NthElement(values, k))
		}
	})

	t.Run("partitions the slice around k", func(t *testing.T) {
		for k := range len(unordered) {
			values := slices.Clone(unordered)
			NthElement(values, k)
			pivot := values[k]
			for _, v := range values[:k] {
				assert.LessOrEqual(t, v, pivot)
			}
			for _, v := range values[k+1:] {
				assert.GreaterOrEqual(t, v, pivot)
			}
		}
	})

	t.Run("reorders the input in place, preserving elements", func(t *testing.T) {
		values := slices.Clone(unordered)
		NthElement(values, 2)
		assert.Equal(t, 3, values[2])
		assert.ElementsMatch(t, unordered, values)
	})

	t.Run("single element slice", func(t *testing.T) {
		assert.Equal(t, 42, NthElement([]int{42}, 0))
	})

	t.Run("handles duplicate values", func(t *testing.T) {
		withDupes := []int{3, 1, 3, 1, 2, 3, 1}
		sorted := slices.Sorted(slices.Values(withDupes))
		for k := range len(withDupes) {
			values := slices.Clone(withDupes)
			assert.Equal(t, sorted[k], NthElement(values, k))
		}
	})

	t.Run("already sorted input", func(t *testing.T) {
		for k := range 5 {
			assert.Equal(t, k+1, NthElement([]int{1, 2, 3, 4, 5}, k))
		}
	})

	t.Run("reverse sorted input", func(t *testing.T) {
		for k := range 5 {
			assert.Equal(t, k+1, NthElement([]int{5, 4, 3, 2, 1}, k))
		}
	})

	t.Run("works with any ordered type", func(t *testing.T) {
		words := []string{"cherry", "apple", "date", "banana"}
		assert.Equal(t, "apple", NthElement(slices.Clone(words), 0))
		assert.Equal(t, "date", NthElement(slices.Clone(words), 3))
	})

	t.Run("large shuffled input, every k", func(t *testing.T) {
		const count = 1000
		rng := rand.New(rand.NewPCG(1, 2))
		shuffled := makeIntegers(count)
		rng.Shuffle(count, func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		for k := range count {
			values := slices.Clone(shuffled)
			require.Equal(t, k+1, NthElement(values, k))
		}
	})

	t.Run("panics on negative k", func(t *testing.T) {
		assert.Panics(t, func() { NthElement([]int{1, 2, 3}, -1) })
	})

	t.Run("panics when k equals length", func(t *testing.T) {
		assert.Panics(t, func() { NthElement([]int{1, 2, 3}, 3) })
	})

	t.Run("panics when k exceeds length", func(t *testing.T) {
		assert.Panics(t, func() { NthElement([]int{1, 2, 3}, 99) })
	})

	t.Run("panics on empty slice", func(t *testing.T) {
		assert.Panics(t, func() { NthElement([]int{}, 0) })
	})
}

func TestNthElementFunc(t *testing.T) {
	t.Parallel()

	t.Run("selects using a descending comparator", func(t *testing.T) {
		descending := func(a, b int) int { return cmp.Compare(b, a) }
		values := []int{5, 3, 1, 4, 2}
		// Under descending order the 0th element is the largest.
		assert.Equal(t, 5, NthElementFunc(slices.Clone(values), 0, descending))
		assert.Equal(t, 1, NthElementFunc(slices.Clone(values), 4, descending))
	})

	t.Run("selects a custom struct type by field", func(t *testing.T) {
		items := []testItem{{3, 30}, {1, 10}, {5, 50}, {2, 20}, {4, 40}}
		got := NthElementFunc(items, 0, testItem.compare)
		assert.Equal(t, testItem{1, 10}, got)
	})

	t.Run("matches NthElement under cmp.Compare", func(t *testing.T) {
		unordered := []int{5, 3, 1, 4, 2}
		for k := range len(unordered) {
			got := NthElementFunc(slices.Clone(unordered), k, cmp.Compare[int])
			assert.Equal(t, k+1, got)
		}
	})

	t.Run("panics on out-of-range k", func(t *testing.T) {
		assert.Panics(t, func() {
			NthElementFunc([]int{1, 2, 3}, 3, cmp.Compare[int])
		})
	})
}

// TestPartitionFunc exercises the unexported partitionFunc step on its own.
//
// lo and hi are the inclusive bounds of the window partitionFunc reorders; the run
// driver supplies them, starting on the full slice and narrowing toward k.
func TestPartitionFunc(t *testing.T) {
	t.Parallel()

	t.Run("upholds the partition invariant", func(t *testing.T) {
		//
		values := []int{9, 3, 7, 1, 8, 2, 5}
		i := partitionFunc(values, 0, len(values)-1, cmp.Compare[int])

		// Values his been modified so that everything left of the pivot
		// is less than the pivot; everything right is greater.
		pivot := values[i]
		assert.Equal(t, 5, pivot)
		expected := []int{1, 3, 2, 5, 8, 7, 9}
		assert.Equal(t, expected, values)

		// Everything left of the pivot is less than the pivot
		for _, v := range values[:i] {
			assert.Less(t, v, pivot)
		}

		// Everything right is equal or greater.
		for _, v := range values[i+1:] {
			assert.GreaterOrEqual(t, v, pivot)
		}
	})

	t.Run("picks the median of three as the pivot", func(t *testing.T) {
		// Pivot = median of first/middle/last = median(9, 1, 5) = 5, parked at its sorted index.
		values := []int{9, 3, 7, 1, 8, 2, 5}
		p := partitionFunc(values, 0, len(values)-1, cmp.Compare[int])
		assert.Equal(t, 5, values[p])
		assert.Equal(t, 3, p) // three values (1, 2, 3) are smaller than the pivot
	})

	t.Run("reorders the window without losing elements", func(t *testing.T) {
		// Partitioning is a permutation: the same multiset comes out, merely reordered.
		values := []int{9, 3, 7, 1, 8, 2, 5}
		partitionFunc(values, 0, len(values)-1, cmp.Compare[int])
		assert.ElementsMatch(t, []int{1, 2, 3, 5, 7, 8, 9}, values)
	})

	t.Run("confines its swaps to the lo..hi window", func(t *testing.T) {
		// A partial window, values[2:9]: partition must leave everything outside [lo, hi] put.
		values := []int{-100, -99, 9, 3, 7, 1, 8, 2, 5, 99, 100}
		lo, hi := 2, 8
		partitionFunc(values, lo, hi, cmp.Compare[int])
		assert.Equal(t, []int{-100, -99}, values[:lo])
		assert.Equal(t, []int{99, 100}, values[hi+1:])
	})

	t.Run("sends pivot-equal elements to the right side", func(t *testing.T) {
		// Strict "< pivot" keeps pivot duplicates out of the left side -- the all-equal O(n^2) trap.
		values := []int{5, 5, 5, 5, 5}
		p := partitionFunc(values, 0, len(values)-1, cmp.Compare[int])
		assert.Equal(t, 0, p)
	})
}

type testItem struct {
	index int
	value int
}

// compare orders items by their index field.
func (i testItem) compare(b testItem) int {
	return cmp.Compare(i.index, b.index)
}

// makeIntegers builds a slice of integers from 1 to count.
func makeIntegers(count int) []int {
	numbers := make([]int, 0, count)
	for i := range count {
		numbers = append(numbers, i+1)
	}
	return numbers
}
