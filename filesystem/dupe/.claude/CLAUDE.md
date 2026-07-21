# dupe - duplicate file scanner

CLI that walks one or more directory roots concurrently, hashes every regular file with
SHA-256, and reports duplicate groups. Per-file hashes are cached in SQLite so unchanged files
are skipped on subsequent runs.

## Architecture

Module `local.dev/monarch`. The engine is the `monarch` library at the repo root; the CLI lives
in `cmd/` (`package main`, imports the library) and will grow into a Cobra multi-command tool.

Library (`monarch`):

- `files.go` - the walk and the hash pipeline. `Collector` walks the roots concurrently (call
  `Walk`, then read `.Folders` / `.AbsRoots`) and emits `FolderInfo` values (path, mtime, child
  names). `Scanner` (`NewScanner(cache, jobs, log, force)`) owns the worker pool;
  `Scanner.Process(folders)` runs the pipeline and returns `[]FileInfo`. `hashFile` is a pure
  SHA-256 helper.
- `cache.go` - SQLite-backed persistent hash cache.
- `report.go` - pure aggregation over `[]FileInfo`: `Summarize`, `ExtensionStats`,
  `DuplicateGroups` return data models, no I/O.

CLI (`cmd/`):

- `main.go` - flag parsing, logger, `cachePath`, wiring.
- `report.go` - presentation: formats the library's report models to an `io.Writer`; owns
  `formatSize`.
- `multihandler.go` - fan-out `slog.Handler` (console + JSON log file).

Every component takes a `*slog.Logger` via its constructor; passing `nil` swaps in a discard
handler. The package-level `slog` default is never used.

## Cache

A `Cache` whose `db == nil` is a no-op (every method short-circuits) - the path taken when
`OpenCache` fails but we still want to run without persistence.

- `Set` is async: a single writer goroutine owns the write transaction and commits every
  `flushSize` rows or `flushInterval`, whichever comes first. `Flush` / `Close` block until
  queued writes land; per-row errors are logged by the writer, not returned to callers.
- `Sweep(folders, roots)` is synchronous: drops cache rows for folders/files no longer seen
  under the given roots, preserving folders outside those roots.
- Reads (`Get`, `GetFolderMtime`, `GetFilesInFolder`) are synchronous.

## SQLite backend

- Driver `modernc.org/sqlite` (pure Go, no cgo), registered as `"sqlite"` not `"sqlite3"`;
  builds with `CGO_ENABLED=0`.
- Schema (`folders` + `files`) lives in `cache.go`, versioned via `PRAGMA user_version`. A
  version mismatch drops and rebuilds the tables, so the next scan re-hashes everything; a
  non-SQLite file at the cache path is deleted and recreated.
- WAL mode leaves `cache.db-wal` / `cache.db-shm` sidecars, cleaned up on clean shutdown.

## Folder-mtime trust

`Scanner.Process` trusts a folder whose cached mtime equals its current mtime (unless
`--force`) and serves every child from cache instead of re-hashing. Files on disk but missing
from cache still go through the worker pool.

This is imperfect - a folder's mtime does not change when a file is edited in place - so
`verifyDuplicates` re-stats only the cached files that appear in candidate duplicate groups and
re-hashes any that drifted. The asymmetry that makes this cheap: an in-place edit moves a hash
randomly, so a stale cached hash colliding with another file is 2⁻²⁵⁶. The real risk is false
positives (re-stat catches them); false negatives are negligible.

## Tuning knobs (`cache.go`)

- `flushSize` (10000 rows) and `flushInterval` (1s) bound the write transaction. The 1s cap
  matches a "happy to lose a second or two on Ctrl-C" tolerance.
- `writeChanBuffer` (64) smooths `Set`; full means backpressure. Worst-case loss on an
  un-graceful exit is roughly `writeChanBuffer + flushSize` rows.

## CLI

`-v`/`-q` logging, `-m` min duplicate size, `-j` worker count (default `NumCPU`), `-f` force
(ignore the folder-mtime cache). See `cmd/main.go` for exact defaults.

## Verification

`go vet ./...` and `go test ./...`. For a full cache rebuild, `rm ~/.cache/dupe/cache.db*`.

## Deferred / out of scope

- Index on `files(hash)` to make duplicate grouping a SQL query (roughly doubles write cost).
- Query helpers (`Cache.Duplicates()` etc.).
- Cross-process locking (WAL allows concurrent opens; only writes serialise).
