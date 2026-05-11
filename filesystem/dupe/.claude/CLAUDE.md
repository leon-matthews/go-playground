# dupe — duplicate file scanner

CLI that walks one or more directory roots concurrently, hashes every regular file with SHA-256,
and reports duplicate groups. Per-file hashes are persisted in a SQLite cache so subsequent runs
are near-instant when files are unchanged.

## Architecture

- `main.go` — flag parsing, logger construction, wires everything together.
- `files.go` — `Scanner` struct owns the worker pool. `newScanner(cache, maxWorkers, log)`
  returns one; `Scanner.Process(paths)` runs the hash pipeline; `Scanner.processFile` is
  unexported. `collectRoot` / `collectRoots` walk the directory trees concurrently. `hashFile`
  is a pure SHA-256 helper.
- `cache.go` — SQLite-backed persistent hash cache (see below).
- `report.go` — pure data processing on `FileInfo` slices: `analyse`, `printSummary`,
  `printByExtension`, `printDuplicates`, `formatSize`.

The package-level `slog` default is not used anywhere. Every component takes a `*slog.Logger`
via constructor; passing `nil` swaps in `slog.New(slog.DiscardHandler)`.

## Cache contract

Public API (`cache.go`):

- `openCache(path string, log *slog.Logger) (*Cache, error)` — creates a SQLite database at
  path. If the file exists and isn't SQLite (magic header mismatch), it's logged + deleted +
  recreated. Empty path returns a no-op cache.
- `Cache.Get(path) (CacheEntry, bool)` — synchronous SELECT, returns hit/miss.
- `Cache.Set(path, entry) error` — **async**. Sends to the writer's channel and returns `nil`
  immediately. Per-row errors are logged by the writer; callers don't see them.
- `Cache.Flush() error` — blocks until all writes queued before this call have committed.
  Useful in tests and at scan checkpoints.
- `Cache.Sweep(seen, roots) error` — **synchronous**. Removes entries under any of `roots`
  that aren't in `seen`. Out-of-scope entries are preserved.
- `Cache.Close() error` — drains the writer, commits any pending tx, closes the DB.
  Idempotent via `sync.Once`.

A `Cache` with `db == nil` is a no-op: every method short-circuits. This is the path taken
when `openCache` errors but the caller still wants to keep running without persistence.

## SQLite backend

- Driver: `modernc.org/sqlite` (pure Go, no cgo). Registered as `"sqlite"`, not `"sqlite3"`.
  Builds work with `CGO_ENABLED=0`.
- DSN bundles pragmas: `journal_mode=WAL`, `synchronous=NORMAL`, `temp_store=MEMORY`,
  `cache_size=-65536` (64 MB).
- Schema:
  ```sql
  CREATE TABLE hashes (
      path     TEXT    PRIMARY KEY,
      size     INTEGER NOT NULL,
      modtime  INTEGER NOT NULL,  -- unix nanoseconds
      hash     BLOB    NOT NULL   -- 32 raw bytes; exposed in Go as [32]byte
  ) WITHOUT ROWID;
  ```
- No indexes (yet). Add them as query needs emerge.
- WAL produces `cache.db-wal` and `cache.db-shm` sidecars alongside `cache.db`. Expected;
  cleaned up on clean shutdown.

## Writer goroutine

One goroutine owns the write transaction. It receives `cacheOp` values through `c.writes`
(buffer 64). Each op is tagged write / flush / sweep:

- **write** — open tx if needed, prepared `INSERT OR REPLACE`. Commits every `flushSize`
  (10000) rows or every `flushInterval` (1s), whichever first.
- **flush** — commit current tx, ack the caller's channel.
- **sweep** — commit current tx, run `runSweep` (temp `seen` table + per-root prefix `DELETE`),
  ack the caller's error channel.

On `Close`: `c.done` is closed; the writer drains any remaining ops from `c.writes` using
`select { default: }`, commits, and exits. `Close` blocks on `wg.Wait()` so all queued work
is persisted.

The prefix-match used by Sweep is `substr(path, 1, length(?)+1) = ? || '/'`, not `LIKE`, so
filenames containing `%` or `_` are handled correctly.

## Tuning knobs

All in `cache.go`:

- `flushSize = 10000` — rows per transaction before forced commit.
- `flushInterval = 1 * time.Second` — wall-clock cap; matches "I'd be happy to lose a second
  or two of data on Ctrl-C" tolerance.
- `writeChanBuffer = 64` — small smoothing buffer; full → backpressure on `Set`. Bump if a
  fast NVMe with many workers ever stalls workers waiting to enqueue.
- Worst-case data loss on un-graceful shutdown ≈ `writeChanBuffer + flushSize` rows.

CLI:
- `-j N` / `--jobs N` — worker pool size (default `runtime.NumCPU()`). Lower on spinning rust
  / NAS where concurrent reads thrash the head; default is fine on SSD.

## Verification

```sh
go vet ./...
go test ./...
```

Smoke test:

```sh
rm -f ~/.cache/dupe/cache.db*
go run . /some/dir       # cold run; logs "starting workers"
go run . /some/dir       # warm run; near-instant
sqlite3 ~/.cache/dupe/cache.db 'SELECT COUNT(*) FROM hashes;'
```

Migration test:

```sh
echo "not-a-db" > ~/.cache/dupe/cache.db
go run . /some/dir       # logs "cache file is not a SQLite database; recreating"
```

## Deferred / out of scope

- **Indexes on the `hashes` table.** First likely add: `CREATE INDEX idx_hash ON hashes(hash)`
  to make `GROUP BY hash HAVING COUNT(*) > 1` near-instant. Roughly doubles write cost, so add
  only when a query needs it.
- **Richer query helpers** (`Cache.Duplicates()`, etc.). The schema supports them; we'll add
  methods as concrete use cases arise.
- **bbolt → SQLite migration of existing cache contents.** Discarded by design; first run
  after upgrade re-hashes.
- **Cross-process locking.** WAL allows concurrent opens; only writes serialise. If accidental
  concurrent runs become a problem, add an advisory lock.
