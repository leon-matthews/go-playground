package quickselect_test

import (
	"cmp"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"quickselect/quickselect"
)

func TestNthElement(t *testing.T) {
	t.Parallel()
	unordered := []int{5, 3, 1, 4, 2}

	t.Run("returns the minimum at k=0", func(t *testing.T) {
		assert.Equal(t, 1, quickselect.NthElement(slices.Clone(unordered), 0))
	})

	t.Run("returns the maximum at k=len-1", func(t *testing.T) {
		assert.Equal(t, 5, quickselect.NthElement(slices.Clone(unordered), len(unordered)-1))
	})

	t.Run("returns the kth smallest for every k", func(t *testing.T) {
		for k := range len(unordered) {
			values := slices.Clone(unordered)
			assert.Equal(t, k+1, quickselect.NthElement(values, k))
		}
	})

	t.Run("partitions the slice around k", func(t *testing.T) {
		for k := range len(unordered) {
			values := slices.Clone(unordered)
			quickselect.NthElement(values, k)
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
		quickselect.NthElement(values, 2)
		assert.Equal(t, 3, values[2])
		assert.ElementsMatch(t, unordered, values)
	})

	t.Run("single element slice", func(t *testing.T) {
		assert.Equal(t, 42, quickselect.NthElement([]int{42}, 0))
	})

	t.Run("handles duplicate values", func(t *testing.T) {
		withDupes := []int{3, 1, 3, 1, 2, 3, 1}
		sorted := slices.Sorted(slices.Values(withDupes))
		for k := range len(withDupes) {
			values := slices.Clone(withDupes)
			assert.Equal(t, sorted[k], quickselect.NthElement(values, k))
		}
	})

	t.Run("already sorted input", func(t *testing.T) {
		for k := range 5 {
			assert.Equal(t, k+1, quickselect.NthElement([]int{1, 2, 3, 4, 5}, k))
		}
	})

	t.Run("reverse sorted input", func(t *testing.T) {
		for k := range 5 {
			assert.Equal(t, k+1, quickselect.NthElement([]int{5, 4, 3, 2, 1}, k))
		}
	})

	t.Run("works with any ordered type", func(t *testing.T) {
		words := []string{"cherry", "apple", "date", "banana"}
		assert.Equal(t, "apple", quickselect.NthElement(slices.Clone(words), 0))
		assert.Equal(t, "date", quickselect.NthElement(slices.Clone(words), 3))
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
			require.Equal(t, k+1, quickselect.NthElement(values, k))
		}
	})

	t.Run("panics on negative k", func(t *testing.T) {
		assert.Panics(t, func() { quickselect.NthElement([]int{1, 2, 3}, -1) })
	})

	t.Run("panics when k equals length", func(t *testing.T) {
		assert.Panics(t, func() { quickselect.NthElement([]int{1, 2, 3}, 3) })
	})

	t.Run("panics when k exceeds length", func(t *testing.T) {
		assert.Panics(t, func() { quickselect.NthElement([]int{1, 2, 3}, 99) })
	})

	t.Run("panics on empty slice", func(t *testing.T) {
		assert.Panics(t, func() { quickselect.NthElement([]int{}, 0) })
	})
}

func TestNthElementCustom(t *testing.T) {
	t.Parallel()

	t.Run("selects using a descending comparator", func(t *testing.T) {
		descending := func(a, b int) int { return cmp.Compare(b, a) }
		values := []int{5, 3, 1, 4, 2}
		// Under descending order the 0th element is the largest.
		assert.Equal(t, 5, quickselect.NthElementCustom(slices.Clone(values), 0, descending))
		assert.Equal(t, 1, quickselect.NthElementCustom(slices.Clone(values), 4, descending))
	})

	t.Run("selects a custom struct type by field", func(t *testing.T) {
		items := []testItem{{3, 30}, {1, 10}, {5, 50}, {2, 20}, {4, 40}}
		got := quickselect.NthElementCustom(items, 0, testItem.compare)
		assert.Equal(t, testItem{1, 10}, got)
	})

	t.Run("matches NthElement under cmp.Compare", func(t *testing.T) {
		unordered := []int{5, 3, 1, 4, 2}
		for k := range len(unordered) {
			got := quickselect.NthElementCustom(slices.Clone(unordered), k, cmp.Compare[int])
			assert.Equal(t, k+1, got)
		}
	})

	t.Run("panics on out-of-range k", func(t *testing.T) {
		assert.Panics(t, func() {
			quickselect.NthElementCustom([]int{1, 2, 3}, 3, cmp.Compare[int])
		})
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
