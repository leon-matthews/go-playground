# pwnedcache

Download, update, and query an offline copy of the
[Have I Been Pwned password database](https://haveibeenpwned.com/Passwords).

The Pwned Passwords corpus holds the SHA-1 hash of every password seen in a public data breach,
along with the number of times each one appeared. `pwnedcache` mirrors the whole corpus into a
local SQLite database, keeps it up to date cheaply, and checks passwords against it without
sending anything over the network.


## Requirements

- Go 1.26 or later, to build. No C compiler is needed: SQLite is provided by the pure-Go
  [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite).
- About 60 GB of free disk space for the database.
- A reasonable internet connection. A full download makes over a million HTTP requests.


## Installation

<!-- TODO: add `go install` instructions once the module path is settled -->

Build from source:

    $ git clone <repository URL>
    $ cd pwnedcache
    $ make build

This produces a single `pwnedcache` binary in the current folder.


## Usage

### Downloading and updating

    $ pwnedcache update

The corpus is split into 1,048,576 hash lists, one for each 5-character hexadecimal prefix of a
SHA-1 hash. The first run downloads every one of them into `pwnedcache.db`, in the current
folder.

Later runs send each list's stored ETag with the request, so only lists that have changed since
the last run are downloaded again. A refresh of a complete database typically fetches a few
thousand changed lists and finishes in a few minutes.

Pressing Ctrl-C stops a run cleanly. Every completed list is already saved, so the next run
picks up where the last one left off.

Options:

| Flag                | Default | Description                                             |
|---------------------|---------|---------------------------------------------------------|
| `-c, --concurrency` | 64      | Number of parallel fetch workers                        |
| `--limit`           | 0       | Stop after this many prefixes (0 means no limit)        |
| `-p, --progress`    | 10s     | Interval between progress reports                       |
| `--retries`         | 10      | Retry attempts per failed fetch (0 disables)            |
| `--profile`         | off     | Write CPU and heap profiles (`cpu.pprof`, `heap.pprof`) |

While most requests are cheap "not modified" checks, `update` runs ten times the requested
number of workers. It drops back to the normal number as soon as the first changed list arrives.

Each run writes a structured (NDJSON) log to `pwnedcache.log` in the current folder, replacing
the log from the previous run.

### Checking passwords

    $ pwnedcache check password123 'correct horse battery staple'
    PASSWORD                      BREACHES
    password123                   2266543
    correct horse battery staple  391

Each password is hashed locally and looked up in the database. A count of zero means that the
password does not appear in the corpus. The database is opened read-only, and nothing is sent
over the network.

**Note:** passwords given on the command line are saved in your shell history, and can be seen
by other users on the same machine while the command runs. Avoid checking passwords you
actually use this way.

### Global options

| Flag             | Default         | Description                 |
|------------------|-----------------|-----------------------------|
| `-d, --database` | `pwnedcache.db` | Path to the SQLite database |
| `-v, --verbose`  | off             | Debug-level logging         |


## Development

    $ make check    # revive, tests, and govulncheck

The database layer is generated from SQL by [sqlc](https://sqlc.dev/); see
[internal/database/README.md](internal/database/README.md).


## Acknowledgements

The password data comes from [Have I Been Pwned](https://haveibeenpwned.com/) by Troy Hunt,
through its free [Pwned Passwords API](https://haveibeenpwned.com/API/v3#PwnedPasswords).


## License

Licensed under the [Apache License, Version 2.0](LICENSE).
