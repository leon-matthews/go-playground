# Mersenne

Uses the arbitrary precision integers in Go's `math/big` package to 'find' Mersenne
primes, using every CPU core on your machine.

## TODO

- Add a user-configurable limit to the max exponent, ie. 1000 to check all prime
  candidates less than to 2^1000-1. A good default might be 5,000

- Set up a pipeline with GOMAXPROCS primecheckers to max out the CPU.
