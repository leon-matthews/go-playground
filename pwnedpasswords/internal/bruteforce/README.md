# bruteforce

Generate candidate passwords by brute force and record any that match the breach corpus.

`Run` opens the output database and the read-only `pwnedcache`, loads the membership
`filter` if one is present, and enumerates every candidate string in odometer order,
shortest first. Each candidate is hashed and checked by a `checker.Checker`; a match is
upserted into the output database with its breach count. With a filter present the work is
sharded across CPUs by `searchParallel`; without one it falls back to the slow serial
`searchSerial`.

## The odometer

A candidate of length _n_ is an `[]int` of _n_ odometer indices into the `alphabet`. All the
enumeration is index arithmetic, converted to bytes only when a candidate is actually built:

- `advance` increments the odometer by one, least-significant digit last, and returns
  `false` on a complete roll-over (the end of a length).
- `addN` jumps the odometer forward by _n_, used to skip a whole chunk at once. It returns
  `false` if the jump rolls past the last candidate of the length.
- `subN` steps the odometer backwards, clamping at all-zeros, used to compute a resume point.
- `pattern` renders indices back into the candidate string, for display and resuming.
- `powSat` computes `base**length` saturated at `math.MaxUint64`, so an astronomically large
  candidate space can be compared against ordinary bounds without overflow.

## Assigning work: the coordinator

Concurrency is pull-based. A single `coord` owns the odometer position `cur` for the current
length and hands out contiguous **chunks** of the space to whichever worker asks next, under
its mutex `mu`:

- `reset(start)` positions the coordinator at the beginning of a new length.
- `next()` returns the start indices of the next chunk and advances `cur` by `chunk`
  candidates with `addN`, setting `done` once the length is exhausted. A worker that gets
  `ok == false` has no more work for this length.

Because a chunk is described only by its start index and the fixed `chunk` size, no candidate
list is ever materialised or copied between goroutines - each worker regenerates its chunk
locally with `advance`. The `parallel chunking covers every candidate exactly once` test
pins down that `next` + local `advance` partitions the space with no gaps or overlaps.

The chunk size is chosen per length by `chunkForSpace`, from the space returned by `powSat`:
it aims for `chunksPerWorker` (8) chunks per worker so uneven per-candidate cost still
balances out, clamped to `[minChunk, maxChunk]`. A short length whose whole space is smaller
than `minChunk` therefore still splits across workers instead of landing entirely on one; a
huge length keeps chunks bounded at `maxChunk` so the coordinator lock stays cheap.

## Doing the work: the workers

For each length, `searchParallel` sets `co.chunk`, calls `co.reset(cur)`, then launches
`workers` goroutines with `wg.Go`, each running `searchWorker`, and blocks on `wg.Wait()`
before moving to the next length. Every worker loops:

1. Pull a chunk start from `co.next()`; stop when the length is done.
2. Copy the start into its own `indices` and step up to `chunk` times, building each
   candidate into a reused `buf` and passing it to `chk.Check`.
3. `advance` the odometer each step, breaking early on roll-over at the end of the length.

Each worker keeps a goroutine-local `progress.Tally`, so `chk.Check` records filter and hash
lookups, matches, and the most recent sample without touching any shared atomic on the hot
path.

## Collecting results

There are two independent collection paths, so the hot loop never serialises on shared state:

- **Matches** are written straight through by the `checker.Checker`, which on a filter hit
  looks the hash up in the cache and calls `Writer.Upsert` on the shared
  `database.BatchWriter`. The writer is the one component all workers genuinely share for
  output, and it batches internally.
- **Counts** are folded from each worker's local `Tally` into the shared
  `progress.Progress` by `prog.Add` every `progress.FlushEvery` candidates, once more at the
  end of each chunk, and a final time via `defer prog.Add(&t)` when the worker exits. `Add`
  updates atomics and resets the local tally. A `progress.Reporter`, started in `Run`, reads
  a `Snapshot` on its own ticker and logs it - it never blocks the workers.

## Cancellation and resume

`searchParallel` derives a cancellable `runCtx`. The first worker to return an error records
it in `firstErr` under a `sync.Once` and calls `cancel`, tearing the others down. Within a
chunk a worker only tests `ctx.Err()` every `ctxCheckMask + 1` candidates, so cancellation is
prompt without a context check on every single candidate.

On a clean interrupt (Ctrl-C, `ctx.Err() != nil`), `searchParallel` logs a resume pattern
computed from `co.frontier(workers)`: since at most `workers` chunks can be in flight,
rewinding `cur` by `workers * chunk` with `subN` gives a point guaranteed to be behind all
processed candidates, so `--resume` never skips one. `resumeStart` turns a `--resume` string
back into a starting length and odometer indices.

## Serial fallback

When no filter is available `Run` uses `searchSerial`: a single loop that runs every
candidate through `chk.Check` (which then hits the database directly), advancing the odometer
and bumping the length on roll-over. It shares the same `Tally`/`FlushEvery` progress
folding but has no coordinator or workers. It exists only as a correctness fallback; without
the filter to skip ~99.9% of lookups it is far too slow for real brute-force search.

## Tuning constants

| Constant | Meaning |
| --- | --- |
| `maxChunk` | Upper bound on a chunk, capping coordinator hand-out frequency. |
| `minChunk` | Lower bound on a chunk, so short lengths still split across workers. |
| `chunksPerWorker` | Target chunks per worker, for load balancing within a length. |
| `ctxCheckMask` | Mask controlling how often a worker checks for cancellation mid-chunk. |
