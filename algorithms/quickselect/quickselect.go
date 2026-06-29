// Package quickselect efficiently finds the k-th smallest element of a slice
// using Tony Hoare's [Quickselect] algorithm.
//
// [Quickselect]: https://en.wikipedia.org/wiki/Quickselect
package quickselect

import (
	"cmp"
	"fmt"
)

// NthElement returns the k-th smallest element (0-based), modifying input.
//
// Only a partial ordering is performed, making it many times faster than pre-sorting
// when only one value is required.
//
// Clone the slice first with [slices.Clone] if you need to keep the original order.
// Panics if k is outside [0, len(values)).
func NthElement[T cmp.Ordered](values []T, k int) T {
	return run(values, k, partition[T])
}

// NthElementFunc is like [NthElement] but orders elements with the given comparator.
//
// cmp reports whether a is less than (negative), equal to (zero), or greater than
// (positive) b, the same convention as [cmp.Compare] and [slices.SortFunc].
func NthElementFunc[T any](values []T, k int, cmp func(T, T) int) T {
	return run(values, k, func(values []T, lo, hi int) int {
		return partitionFunc(values, lo, hi, cmp)
	})
}

// run narrows the active window until a pivot lands on k, delegating each pivot step to
// partitionStep. Descending into only the side that holds k is what turns the average
// cost from O(n log n) into O(n).
func run[T any](values []T, k int, partitionStep func(values []T, lo, hi int) int) T {
	if k < 0 || k >= len(values) {
		panic(fmt.Sprintf("quickselect: k=%d out of range [0, %d)", k, len(values)))
	}
	lo, hi := 0, len(values)-1
	for lo < hi {
		p := partitionStep(values, lo, hi)
		switch {
		case k < p:
			hi = p - 1
		case k > p:
			lo = p + 1
		default:
			return values[k]
		}
	}
	return values[k]
}

// partition is [partitionFunc] for basic ordered types.
//
// Comparing with the < operator inlines and avoids the indirect comparator call that
// dominates NthElementFunc for cheap element types.
//
// Interestingly, the Lomuto scheme was first chosen for being easy to understand,
// but an experiment switching it out to Hoare's original method showed that to
// be slower on all input, except for the all-values-equal case, presumably
// for locality of reference reasons.
func partition[T cmp.Ordered](values []T, lo, hi int) int {
	medianOfThree(values, lo, hi)

	// Lomuto scheme: grow a "less than pivot" prefix as j sweeps the window.
	pivot := values[hi]
	i := lo
	for j := lo; j < hi; j++ {
		if values[j] < pivot {
			values[i], values[j] = values[j], values[i]
			i++
		}
	}
	values[i], values[hi] = values[hi], values[i]
	return i
}

// medianOfThree is [medianOfThreeFunc] using the < operator; see [partition].
func medianOfThree[T cmp.Ordered](values []T, lo, hi int) {
	mid := lo + (hi-lo)/2
	if values[mid] < values[lo] {
		values[mid], values[lo] = values[lo], values[mid]
	}
	if values[hi] < values[lo] {
		values[hi], values[lo] = values[lo], values[hi]
	}
	if values[hi] < values[mid] {
		values[hi], values[mid] = values[mid], values[hi]
	}
	// Now values[lo] <= values[mid] <= values[hi]; move the median to hi as the pivot.
	values[mid], values[hi] = values[hi], values[mid]
}

// partitionFunc reorders values[lo:hi+1] around a pivot and returns the pivot's final index.
// Everything left of that index is less than the pivot; everything right is >= it.
func partitionFunc[T any](values []T, lo, hi int, cmp func(T, T) int) int {
	medianOfThreeFunc(values, lo, hi, cmp)

	// Lomuto scheme: grow a "less than pivot" prefix as j sweeps the window.
	pivot := values[hi]
	i := lo
	for j := lo; j < hi; j++ {
		if cmp(values[j], pivot) < 0 {
			values[i], values[j] = values[j], values[i]
			i++
		}
	}
	values[i], values[hi] = values[hi], values[i]
	return i
}

// medianOfThreeFunc parks the median of the first, middle, and last elements at hi.
// Choosing the median as pivot keeps sorted and reverse-sorted inputs off the O(n^2)
// worst case that a naive first- or last-element pivot would hit.
func medianOfThreeFunc[T any](values []T, lo, hi int, cmp func(T, T) int) {
	mid := lo + (hi-lo)/2
	if cmp(values[mid], values[lo]) < 0 {
		values[mid], values[lo] = values[lo], values[mid]
	}
	if cmp(values[hi], values[lo]) < 0 {
		values[hi], values[lo] = values[lo], values[hi]
	}
	if cmp(values[hi], values[mid]) < 0 {
		values[hi], values[mid] = values[mid], values[hi]
	}
	// Now values[lo] <= values[mid] <= values[hi]; move the median to hi as the pivot.
	values[mid], values[hi] = values[hi], values[mid]
}
