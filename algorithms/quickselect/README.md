# Quickselect

Generic quickselect implementation: finds the k-th smallest element of a slice in O(n)
average time, without fully sorting it.

`NthElement` works on any `cmp.Ordered` type; `NthElementFunc` takes a comparator. Both
reorder the slice in place and panic if `k` is out of range.

## TODO

1. Explore introselect: fall back to median-of-medians pivot selection when the
   median-of-three pivot keeps splitting badly, to guarantee O(n) worst case rather than
   the current O(n²) worst case.
