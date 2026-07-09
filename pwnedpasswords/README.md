# pwnedpasswords

Build breach-frequency password denylists from word lists.

A small security-research tool. It takes candidate passwords - either from word lists or by
brute-force generation - hashes each one with SHA-1, and looks the hash up in the read-only
`hashes` table of a sibling [pwnedcache](../pwnedcache/) database (an offline copy of the
[Have I Been Pwned password database](https://haveibeenpwned.com/Passwords)). Matches are
stored in `pwnedpasswords.db` with their breach counts, and exported as plain-text or JSON
denylists for use during website account creation. The whole table can also be dumped to CSV
and reloaded with `merge`, which is how the database is backed up or rebuilt under a new schema.

## Prerequisites

A `pwnedcache.db` built by the sibling `pwnedcache` tool. By default it is read from the
current directory; override with `--pwnedcache`.

For `bruteforce`, first build a membership filter (see below). Without it every candidate
hits the database, which is far too slow for brute-force search.

## Usage

Import one or more word lists, recording the passwords found in the breach corpus:

```
pwnedpasswords import rockyou.txt other-list.txt
```

Build the membership filter once from the pwnedcache hashes. It is a split-block Bloom
filter, memory-mapped read-only at query time, that lets `bruteforce` skip ~99.9% of
database lookups. Rebuild it whenever `pwnedcache.db` changes (it will not overwrite an
existing filter, so remove the old one first):

```
pwnedpasswords buildfilter                    # 8 GiB filter (default) -> pwnedpasswords.filter
pwnedpasswords buildfilter --4GB              # smaller filter, higher false-positive rate
```

Generate candidates by brute force in odometer order, shortest first, recording matches as
it goes. It runs in parallel across all CPUs until interrupted with Ctrl-C, which prints the
next pattern to try so the run can be continued later with `--resume`:

```
pwnedpasswords bruteforce --alphabet 1        # lowercase only
pwnedpasswords bruteforce --resume "dfxx"     # continue from this pattern
```

Export the most-breached passwords, ordered by breach count:

```
pwnedpasswords export --top 1000              # plain text, one password per line
pwnedpasswords export --top 1000 --format json
```

Dump the whole table as CSV, or merge a CSV back in - skipping any password already present,
leaving existing counts untouched. Together they back up the database and rebuild it under the
current schema: export from the old file, then merge into a fresh one.

```
pwnedpasswords -d old.db export --format csv > backup.csv
pwnedpasswords -d new.db merge backup.csv
```

`import`, `bruteforce`, and `merge` are all additive: running them again accumulates matches.

## Flags

Persistent flags, valid on every command:

- `-d`, `--database` - output SQLite database (default `pwnedpasswords.db`)
- `-c`, `--pwnedcache` - read-only pwnedcache database (default `pwnedcache.db`)
- `-v`, `--verbose` - debug-level logging
- `-q`, `--quiet` - warnings and errors only

`buildfilter` flags. The three size flags are mutually exclusive; the probe count is tuned
per size for the lowest false-positive rate on the ~2 billion hash corpus:

- `--4GB` - 4 GiB filter, false positives ~1 in 1,500
- `--8GB` - 8 GiB filter (default), false positives ~1 in 270,000
- `--16GB` - 16 GiB filter, false positives ~1 in 175 million
- `--filter` - output filter path (default `pwnedpasswords.filter`)
- `-p`, `--progress` - interval between progress reports (default 10s)

`bruteforce` flags:

- `-a`, `--alphabet` - cumulative character set (default 4):
  1 = lowercase, 2 = +space +digits, 3 = +uppercase, 4 = +symbols
- `--resume` - continue from this pattern (as logged when interrupted)
- `--filter` - membership filter path (default `pwnedpasswords.filter`)
- `-w`, `--workers` - number of parallel workers (default: number of CPUs)
- `-p`, `--progress` - interval between progress reports (default 10s)

`export` flags:

- `-n`, `--top` - number of passwords to write (default 1000; `csv` dumps the whole table)
- `-f`, `--format` - output format, `text`, `json`, or `csv` (default `text`)
- `-p`, `--progress` - interval between progress reports (default 10s; `csv` only)

`merge` flags:

- `-p`, `--progress` - interval between progress reports (default 10s)


## Code layout

The tool is split into focused packages. `cmd/` holds only the CLI wiring - a cobra
builder per sub-command that parses flags and calls into a library package. The real work
lives in those packages:

- `checker/` - the shared core: hash a candidate, consult the filter, look any hit up in
  the `hashes` table, and record a match in the output database.
- `search/` - the parallel brute-force engine: odometer candidate enumeration, worker
  coordination, and chunk sizing.
- `wordlist/` - streams word-list files through the checker.
- `buildfilter/` - scans the pwnedcache hashes into the Bloom filter and writes it out.
- `export/` - writes denylists (text, JSON, CSV) and merges a CSV dump back in.
- `filter/` - the split-block Bloom filter: the in-memory structure plus its on-disk,
  memory-mapped format.
- `database/` - opens the read-only pwnedcache and writable output SQLite stores and
  batches password upserts; `database/sqlite/` holds the sqlc-generated queries.
- `progress/` - running totals shared by the scanning commands, and the periodic
  progress/summary reporter.
- `logging/` - the dual console-and-file logger setup.

## Bloom Filter

Database lookups are the bottleneck for bruteforce guessing, where the overwhelming
majority of candidates are misses. I have to use a database, as I can't fit the 
full 50GB hash list in RAM, but we can vastly reduce the number of queries that
we pass on to the database using a Bloom Filter.

A bloom filter is a fascinating *probabistic* data structure that trades huge
space savings for a lack of certainty. Its weakness is false-positives: it is
*always* correct when it says that a key is absent, but sometimes wrong when
it says the key is present.

The best part is that you can tune the filter to get more certainty by spending 
more bytes. For this application we have built support for three different
filter sizes:

| size              | num hashes | false positives   |
|-------------------|------------|-------------------|
| `--4GB`           | 10         | ~1 in 1,500       |
| `--8GB` (default) | 16         | ~1 in 270,000     |
| `--16GB`          | 21         | ~1 in 175 million |

Choose a size that fits comfortably into your system's RAM, while leaving plenty 
free to allow for file-system caching of the input and output database files.
