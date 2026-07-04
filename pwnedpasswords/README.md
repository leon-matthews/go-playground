# pwnedpasswords

Build breach-frequency password denylists from word lists.

A small security-research tool. It takes candidate passwords - either from word lists or by
brute-force generation - hashes each one with SHA-1, and looks the hash up in the read-only
`hashes` table of a sibling [pwnedcache](../pwnedcache/) database (an offline copy of the
[Have I Been Pwned password database](https://haveibeenpwned.com/Passwords)). Matches are
stored in `pwnedpasswords.db` with their breach counts, and exported as plain-text or JSON
denylists for use during website account creation.

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
database lookups. Rebuild it whenever `pwnedcache.db` changes:

```
pwnedpasswords buildfilter                    # ~16 GiB filter -> pwnedpasswords.filter
pwnedpasswords buildfilter --size 8           # smaller filter, higher false-positive rate
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

Both `import` and `bruteforce` are additive: running them again accumulates matches.

## Flags

Persistent flags, valid on every command:

- `-d`, `--database` - output SQLite database (default `pwnedpasswords.db`)
- `-c`, `--pwnedcache` - read-only pwnedcache database (default `pwnedcache.db`)
- `-v`, `--verbose` - debug-level logging
- `-q`, `--quiet` - warnings and errors only

`buildfilter` flags:

- `-s`, `--size` - target filter size in GiB (default 16, rounded down to a power-of-two
  block count)
- `--filter` - output filter path (default `pwnedpasswords.filter`)

`bruteforce` flags:

- `-a`, `--alphabet` - cumulative character set (default 4):
  1 = lowercase, 2 = +space +digits, 3 = +uppercase, 4 = +symbols
- `--resume` - continue from this pattern (as logged when interrupted)
- `--filter` - membership filter path (default `pwnedpasswords.filter`)
- `-w`, `--workers` - number of parallel workers (default: number of CPUs)

`export` flags:

- `-n`, `--top` - number of passwords to write (default 1000)
- `-f`, `--format` - output format, `text` or `json` (default `text`)
