// Package quickselect finds the k-th smallest element of a slice in O(n) average time.
package quickselect

import (
	"cmp"
	"fmt"
)

// NthElement returns the k-th smallest element (0-based), reordering values in place.
//
// After it returns, values[k] holds that element, every item in values[:k] is <= it,
// and every item in values[k+1:] is >= it. Panics if k is outside [0, len(values)).
// Clone the slice first with slices.Clone if you need to keep the original order.
func NthElement[T cmp.Ordered](values []T, k int) T {
	return NthElementFunc(values, k, cmp.Compare[T])
}

// NthElementFunc is like [NthElement] but orders elements with the given comparator.
//
// cmp reports whether a is less than (negative), equal to (zero), or greater than
// (positive) b, the same convention as cmp.Compare and slices.SortFunc.
func NthElementFunc[T any](values []T, k int, cmp func(T, T) int) T {
	if k < 0 || k >= len(values) {
		panic(fmt.Sprintf("quickselect: k=%d out of range [0, %d)", k, len(values)))
	}

	// Narrow the active window [lo, hi] until a pivot lands exactly on k. Unlike
	// quicksort we descend into only the side that contains k, which is what turns
	// the average cost from O(n log n) into O(n).
	lo, hi := 0, len(values)-1
	for lo < hi {
		p := partition(values, lo, hi, cmp)
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

// partition reorders values[lo:hi+1] around a pivot and returns the pivot's final index.
// Everything left of that index is less than the pivot; everything right is >= it.
func partition[T any](values []T, lo, hi int, cmp func(T, T) int) int {
	medianOfThree(values, lo, hi, cmp)

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

// medianOfThree parks the median of the first, middle, and last elements at hi.
// Choosing the median as pivot keeps sorted and reverse-sorted inputs off the O(n^2)
// worst case that a naive first- or last-element pivot would hit.
func medianOfThree[T any](values []T, lo, hi int, cmp func(T, T) int) {
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
