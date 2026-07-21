# mimicry - duplicate file scanner

`mimicry` walks one or more directory trees, hashes every regular file with SHA-256, and
reports groups of files that share identical content. Per-file hashes are cached in SQLite,
so unchanged files are skipped on later runs.

Two stages:

- `mimicry scan ROOT...` walks the roots, hashes files, and populates the cache.
- `mimicry report` reads the cache and prints a summary, a per-extension breakdown, and the
  duplicate groups (largest first).

The cache lives under your user cache directory (e.g. `~/.cache/mimicry/cache.db`); both
commands print its path. Run either command with `--help` for the available flags.
