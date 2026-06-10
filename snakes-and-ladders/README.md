# Go Ladders

A Go port of the Python original, `../snakes_and_ladders.py` — a silly benchmark which
plays many, many solo games of snakes and ladders.

The command-line interface matches the Python script, but the internals are idiomatic Go:
worker goroutines replace the process pool, the board lives in an array rather than a
dictionary, and each worker rolls its dice with its own PCG random number generator from
`math/rand/v2`.

## Build

    go build

## Usage

Play for roughly ten seconds (the default) using a single goroutine:

    ./go_ladders

Play one million games on every core, then dump detailed results to stdout as JSON:

    ./go_ladders -n 1e6 -j --json

Summaries are printed to stderr, so the JSON on stdout can be piped or redirected cleanly.

The only departure from the Python interface is that pflag requires a core count to be
attached to its flag, eg. `-j=4` or `-j4` rather than `-j 4`. A bare `-j` still means
"use every core", as before.
