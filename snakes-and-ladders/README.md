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

Play for ten seconds (the default) using a single goroutine:

    ./go_ladders

Spread one million games across every core, then dump detailed results to stdout as JSON:

    ./go_ladders -n 1_000_000 -j --json

Summaries are printed to stderr, so the JSON on stdout can be piped or redirected cleanly.

The only departure from the Python interface is that the game count is a plain integer,
so exponent notation like `1e6` is not accepted, although Go literal forms such as
`1_000_000` and `0x10` are. The `-j` flag follows the example set by `make`: a bare `-j`
uses every core, while `-j4`, `-j 4`, and `-j=4` all set a count.
