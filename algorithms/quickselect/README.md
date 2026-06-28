# Quickselect

Generic quickselect implementation: finds the k-th smallest element of a slice in O(n)
average time, without fully sorting it.

`NthElement` works on any `cmp.Ordered` type; `NthElementFunc` takes a comparator. Both
reorder the slice in place and panic if `k` is out of range.

## Partitioning: Lomuto vs Hoare

The package partitions with the Lomuto scheme. A benchmark experiment compared it against
classic Hoare two-pointer partitioning, to check whether Hoare's well-known edge in C++
carries over to Go. For the common case it does not:

| Input (median select)       | Hoare vs Lomuto |
| --------------------------- | --------------- |
| distinct ints, n=1e6        | ~6% slower      |
| large struct (512 B), n=5e4 | ~37% slower     |
| all-equal ints, n=1e4       | ~1600× faster   |

Hoare does fewer comparisons and ~4x fewer swaps, yet loses on ordinary inputs: Lomuto's
sequential forward scan with local swaps is far friendlier to the cache and prefetcher,
and the gap grows with element size (a memory-locality bound, not an instruction-count
one). Hoare wins only on duplicate-heavy input, where its stop-on-equal scan sidesteps
Lomuto's O(n²) degradation — an argument for three-way (Dutch-flag) partitioning rather
than for switching schemes wholesale.

## TODO

1. Explore introselect: fall back to median-of-medians pivot selection when the
   median-of-three pivot keeps splitting badly, to guarantee O(n) worst case rather than
   the current O(n²) worst case.
