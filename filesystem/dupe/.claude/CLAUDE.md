# dupe — duplicate file scanner

CLI that walks one or more directory roots concurrently, hashes every regular file with SHA-256,
and reports duplicate groups. Per-file hashes are persisted in a SQLite cache so subsequent runs
are near-instant when files are unchanged.

## Architecture

- `main.go` — flag parsing, logger construction, wires everything together.
- `files.go` — `Scanner` owns the worker pool. `newScanner(cache, maxWorkers, log, force)`
  returns one; `Scanner.Process(folderScans, looseFiles)` runs the hash pipeline.
  `collectRoot` / `collectRoots` walk concurrently and emit `FolderScan` values (path, mtime,
  child basenames) plus a slice of "loose" paths (roots that are themselves regular files).
  `hashFile` is a pure SHA-256 helper. `Scanner.verifyDuplicates` re-stats files from cache
  that appear in candidate duplicate groups (see Folder-mtime trust below).
- `cache.go` — SQLite-backed persistent hash cache (see below).
- `report.go` — pure data processing on `FileInfo` slices.

The package-level `slog` default is not used anywhere. Every component takes a `*slog.Logger`
via constructor; passing `nil` swaps in `slog.New(slog.DiscardHandler)`.

## Cache contract

Public API (`cache.go`):

- `openCache(path string, log *slog.Logger) (*Cache, error)` — creates a SQLite database at
  path. If the file exists and isn't SQLite (magic header mismatch), it's logged + deleted +
  recreated. If the schema version doesn't match the current `schemaVersion`, the old tables
  are dropped and recreated (logged as a warning). Empty path returns a no-op cache.
- `Cache.Get(path) (CacheEntry, bool)` — synchronous SELECT; internally splits path into
  (folder, basename) and does a JOIN against `folders`+`files`.
- `Cache.Set(path, entry) error` — **async**. The writer goroutine resolves-or-creates the
  folder row, then upserts the file row. Per-row errors are logged by the writer; callers
  don't see them.
- `Cache.GetFolderMtime(folder) (time.Time, bool)` — synchronous; for the Scanner's trust check.
- `Cache.SetFolderMtime(folder, mtime) error` — async upsert of the folder's mtime.
- `Cache.GetFilesInFolder(folder) (map[string]CacheEntry, error)` — synchronous bulk fetch,
  keyed by basename. Used on the trusted path.
- `Cache.Flush() error` — blocks until queued writes have committed.
- `Cache.Sweep(seenFolders, seenFiles, roots) error` — **synchronous**. Drops folder rows
  under any root not in `seenFolders` (CASCADE drops their files); then drops files inside
  seen folders whose basename isn't in this folder's seen set. Out-of-scope folders preserved.
- `Cache.Close() error` — drains the writer, commits any pending tx, closes the DB.
  Idempotent via `sync.Once`.

A `Cache` with `db == nil` is a no-op: every method short-circuits. This is the path taken
when `openCache` errors but the caller still wants to keep running without persistence.

## SQLite backend

- Driver: `modernc.org/sqlite` (pure Go, no cgo). Registered as `"sqlite"`, not `"sqlite3"`.
  Builds work with `CGO_ENABLED=0`.
- DSN bundles pragmas: `journal_mode=WAL`, `synchronous=NORMAL`, `temp_store=MEMORY`,
  `cache_size=-65536` (64 MB), `foreign_keys=on` (required for `ON DELETE CASCADE`).
- Schema (`schemaVersion = 2`):
  ```sql
  CREATE TABLE folders (
      id    INTEGER PRIMARY KEY,
      path  TEXT    UNIQUE NOT NULL,
      mtime INTEGER NOT NULL          -- unix nanoseconds; 0 = "seen but mtime unknown"
  );

  CREATE TABLE files (
      folder_id INTEGER NOT NULL,
      name      TEXT    NOT NULL,
      size      INTEGER NOT NULL,
      modtime   INTEGER NOT NULL,
      hash      BLOB    NOT NULL,    -- 32 raw bytes; exposed in Go as [32]byte
      PRIMARY KEY (folder_id, name),
      FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE
  ) WITHOUT ROWID;
  ```
- `PRAGMA user_version = 2` tracks the schema version. On open, a mismatched version triggers
  `DROP TABLE IF EXISTS hashes/files/folders` followed by a fresh `CREATE`. First scan after
  upgrade re-hashes everything.
- No indexes on the user-visible columns yet. Add them when a concrete query needs one
  (`CREATE INDEX idx_files_hash ON files(hash)` is the first likely add).
- WAL produces `cache.db-wal` and `cache.db-shm` sidecars alongside `cache.db`. Expected;
  cleaned up on clean shutdown.

## Writer goroutine

One goroutine owns the write transaction. It receives `cacheOp` values through `c.writes`
(buffer 64). Each op is tagged write / flush / sweep / setFolderMtime.

Per-tx prepared statements:
- `folderEnsureStmt` — `INSERT INTO folders(path, mtime) VALUES (?, 0) ON CONFLICT(path) DO
  UPDATE SET path = excluded.path RETURNING id`. Idempotent "give me the folder id, create
  with mtime=0 if missing, don't clobber an existing mtime".
- `folderMtimeStmt` — same shape but `DO UPDATE SET mtime = excluded.mtime`. Used by
  `SetFolderMtime` to publish the real mtime.
- `fileStmt` — `INSERT INTO files(...) ON CONFLICT(folder_id, name) DO UPDATE SET ...`.

The writer maintains an in-memory `idCache map[string]int64` (folder path → id) to amortize
folder lookups across many file writes. Cleared on rollback or after `Sweep` to stay
consistent with the on-disk state.

Op handling:
- **write** — split path, ensure folder id (via `idCache` or `folderEnsureStmt`), upsert file
  row. Commits every `flushSize` (10000) rows or every `flushInterval` (1s).
- **setFolderMtime** — execute `folderMtimeStmt`; refresh `idCache`.
- **flush** — commit current tx, ack the caller's channel.
- **sweep** — commit current tx, run `runSweep`, ack the caller's error channel. Two-step:
  drop orphan folders under any root (CASCADE drops files); then drop files inside seen
  folders whose basename isn't in the per-folder seen set.

On `Close`: `c.done` is closed; the writer drains any remaining ops using `select { default:
}`, commits, and exits. `Close` blocks on `wg.Wait()` so all queued work is persisted.

## Folder-mtime trust

`Scanner.Process` consults `Cache.GetFolderMtime` on each `FolderScan`:

- **Trusted** (cached mtime equals current folder mtime, and `--force` is off): bulk-fetch
  every cached file in the folder via `GetFilesInFolder`. For each child name in the walk,
  emit a `FileInfo` from the cache entry. Files on disk but missing from cache fall through
  to the per-file worker pool.
- **Stale** (mismatch, or `--force`): dispatch each child to the worker pool; after dispatch,
  queue `SetFolderMtime` so the next scan can trust this folder.

The trust model is imperfect: a folder's mtime does **not** change when an existing file is
modified in place (no `creat`/`unlink`/`rename`), so trusting the mtime can mask in-place
edits. The fix is the verification pass.

### Duplicate verification

After collecting all `FileInfo`s, `Scanner.verifyDuplicates` groups by hash. For any group of
size ≥ 2 with members that came from cache (`fromCache[path] == true`), each member is
re-stat'd. If size or mtime drifted, the file is re-hashed inline and the new entry pushed
back to cache. On stat or hash failure, the file is logged at warn level and dropped from
the result.

The asymmetry that makes this cheap: an in-place edit moves a hash in a random direction,
so the chance of an edited file colliding with another file's hash is 2⁻²⁵⁶. So the
realistic risk is **false positives** (we'd claim two files duplicate when one has drifted);
**false negatives** (real duplicates we miss because a cached hash is stale) require an
edit collision and are negligible. Verification catches the false positives by re-stat'ing
just the files in candidate groups, which is a tiny fraction of total files.

## Tuning knobs

All in `cache.go`:

- `flushSize = 10000` — rows per transaction before forced commit.
- `flushInterval = 1 * time.Second` — wall-clock cap; matches "I'd be happy to lose a second
  or two of data on Ctrl-C" tolerance.
- `writeChanBuffer = 64` — small smoothing buffer; full → backpressure on `Set`. Bump if a
  fast NVMe with many workers ever stalls workers waiting to enqueue.
- Worst-case data loss on un-graceful shutdown ≈ `writeChanBuffer + flushSize` rows.

## CLI

- `-j N` / `--jobs N` — worker pool size (default `runtime.NumCPU()`). Lower on spinning rust
  / NAS where concurrent reads thrash the head; default is fine on SSD.
- `-f` / `--force` — stat every file, ignoring the folder-mtime cache. File-level hash trust
  (size+mtime match) is unchanged. Use this when you've been editing files in place and
  want to re-verify the tree. For a full rebuild, `rm ~/.cache/dupe/cache.db*`.

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
sqlite3 ~/.cache/dupe/cache.db "SELECT COUNT(*) FROM files; SELECT COUNT(*) FROM folders;"
```

Schema migration test:

```sh
sqlite3 ~/.cache/dupe/cache.db "DROP TABLE files; DROP TABLE folders; \
    CREATE TABLE hashes(path TEXT PRIMARY KEY, size INTEGER, modtime INTEGER, hash BLOB); \
    PRAGMA user_version = 1;"
go run . /some/dir       # logs "cache schema version mismatch; rebuilding"
```

Non-SQLite file test:

```sh
echo "not-a-db" > ~/.cache/dupe/cache.db
go run . /some/dir       # logs "cache file is not a SQLite database; recreating"
```

In-place edit test:

```sh
# in a tree with a known duplicate pair
echo "different" > /path/to/one_of_the_dupes
go run . /some/dir       # report should no longer list the modified file as a duplicate
```

## Deferred / out of scope

- **Indexes on the `files` table.** First likely add: `CREATE INDEX idx_files_hash ON
  files(hash)` to make `GROUP BY hash HAVING COUNT(*) > 1` near-instant. Roughly doubles
  write cost, so add only when a query needs it.
- **Richer query helpers** (`Cache.Duplicates()`, etc.). The schema supports them; we'll add
  methods as concrete use cases arise.
- **Cross-process locking.** WAL allows concurrent opens; only writes serialise. If
  accidental concurrent runs become a problem, add an advisory lock.
