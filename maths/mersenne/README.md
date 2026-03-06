# Mersenne

Uses the arbitrary precision integers in Go's `math/big` package to 'find' Mersenne
primes, using every CPU core on your machine.

Just an exercise in using Go's currenency primitives to create a pipeline.

## TODO

- Set up a pipeline with GOMAXPROCS primecheckers to max out the CPU.
