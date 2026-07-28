# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this
repository.

`cine` builds a local SQLite database from the IMDb non-commercial datasets - seven gzipped
TSV files fetched with the `wget` line in `README.md`. A real build takes minutes and produces
a file of some gigabytes, so most work is done against the small TSV fixtures in
`testdata/imdb`. `README.md` documents the flags.

## Commands

Run everything from this folder (`cine/`), which holds `go.mod`.

- `make test` - `gotestsum` with testdox output and `--failfast`
- `make lint` - `revive ./...`
- `go vet ./...` then `revive ./...` then `gofumpt --extra -w .` to validate changes
- One test: `go test ./importer -run TestBuildMetadata` (subtests via `-run 'Test/sub'`)
- Benchmark: `go test ./importer -run '^$' -bench BenchmarkInsertChunkSizes`
- Fuzz: `go test ./reader -run '^$' -fuzz FuzzParseCharacters`
- Regenerate sqlc code: `sqlc generate` from `database/` (config paths are relative to it)
- Real build: `./cine build-database ~/Temp/imdb cine.db`, which is titles-only and filtered
  by default; add `--people` for every table and `--profile` to write `cpu.pprof`
- `./cine reader-benchmark <folder>` reads every record and reports throughput, for tuning
  the reader without touching SQLite

## Architecture

Three packages in a straight line, plus a thin cobra layer:

`reader` -> `importer` -> `database`

- **`reader`** is a faithful, allocation-conscious view of the TSVs. One `ReadX` function
  per file, each returning `iter.Seq2[Record, error]`. Headers are checked against the
  expected columns so a dataset layout change fails loudly. Records keep IMDb's own shapes:
  `tt`/`nm` ids stay strings, `\N` becomes `""` or the `Missing` (-1) sentinel for ints. A
  bad row yields an error carrying its line number without ending iteration.
- **`importer`** owns every transformation. Ids lose their two-letter prefix and become
  integers (`parseID`), ratings become tenths, `\N` becomes SQL NULL, and free-text or
  enumerated fields are interned. `Import` builds into `path + ".tmp"` and renames on
  success, so a failed run never damages an existing database. Each dataset file is one
  "layer" imported in its own transaction, from the slice in `buildInto` - the titles files
  first, then the people ones if `--people` asked for them. `title.ratings` has no table: it
  is loaded into a map first and joined onto titles during their pass, and its only record is
  its `build_sources` row.
- **`database`** embeds both schema files and applies them on `Open` - `schema.sql` always,
  `schema-people.sql` only when its `people` argument says so - and tunes the connection for
  bulk loading (no journal, no fsync, exclusive lock, one connection). `Open` is
  write-focused; a read path should open the finished file separately with gentler pragmas.
- **`cmd`** is cobra wiring only - argument checking, the logger, the `--profile` flag.

### Conventions worth knowing before editing

- **`schema.sql` is the design document.** Every non-obvious column choice is justified in a
  comment there or in `schema-people.sql` with measured counts from a real dump, including
  why IMDb identifier columns carry no foreign key and what a filtered build does to the
  orphans that causes. Read it before changing a table, and read these columns with
  `LEFT JOIN`. `schema.sql` carries the header both files answer to.
- **Bitmask vs junction is a measured decision, not a preference.** A list field becomes a
  bitmask when its order carries nothing (genres, akas types) and a junction with a 1-based
  `position` when IMDb's order is content (primary professions, known-for titles, crew).
  The `interner` serves both: `id()` for lookup foreign keys, `bit()` for bitmask positions.
- **A singular table name means a lookup**, prefixed with the table it serves, so it ends up
  named for the column that references it - `akas.region` resolves through `akas_region`.
  Everything else is plural. That is what tells `akas_attribute` (lookup) from
  `akas_carry_attributes` (junction), and it is stated in `schema.sql`'s header; a new
  lookup that breaks the rule quietly costs the reader that distinction.
- **Bulk inserts are hand-written, not sqlc**, because sqlc cannot emit multi-row SQLite
  inserts. `batchInserter` chunks rows to `bindParamBudget`, a figure measured by
  `BenchmarkInsertChunkSizes` against the modernc driver - read its comment in `batch.go`
  before touching it, and don't raise it on intuition. `database/sqlite/*.go` is generated;
  edit `query.sql` and regenerate instead.
- **Each import pass follows the same shape**: `xColumns` slice, a `bindXRow` function that
  appends values in that order, a `buildXRow` transformer, and an `importX` streaming loop
  returning `counts{read, written}` plus any interners for the caller to flush.
- **"Layer" means one dataset file in one transaction, and nothing else.** It used to also
  name a group of tables in `schema.sql`, which now says Titles and People instead, so the
  word is the importer's alone: `type layer`, `importTitlesLayer` and its five siblings.

### Build options

`BuildOptions` in `importer/importer.go` holds the three choices a build makes, one field per
`build_info` column, so a database cannot claim an option that did not apply. Two axes hide in
those three fields: the filters choose which *rows* are kept, `People` chooses which *tables*
exist at all. `cmd` inverts the two filter flags once each, because `--allow-adult` and
`--allow-unrated` name what to keep while the fields name what to restrict; `--people` names
what to add and so inverts nowhere.

### Titles and people

A table belongs to Titles if it can be populated without knowing that any person exists.
That sorts every table without argument, and - the part that makes it cheap - it partitions
the seven source files with nothing straddling the boundary, so `buildInto` gates whole
layers and no pass ever runs partially or needs to know a layer exists.

The split runs through the schema too: the People tables live in `schema-people.sql`, applied
only when they will be filled, so a build creates nothing it leaves empty. That works because
no Titles table references a People one and the People tables' only foreign key parents are
their own lookups - the `title_id` columns that cross the line carry no foreign key by design.
The cost is that `sqlite.Prepare` cannot be used, since preparing `ListPrincipalsForTitle`
resolves a table a titles-only database has not got; `Open` returns `sqlite.New(db)` instead,
which the import path does not mind because it never uses the generated queries.

Ordering the layers titles-first is presentational, not a dependency: no pass reads a table
another wrote, no two passes share an interner, and both shared inputs - the ratings map and
the filter - are settled before the first layer runs. So the layers may be reordered freely,
which is worth knowing before assuming otherwise from the slice.

### The row filter

`importer/filter.go` builds the allow-list for a smaller second database, from the rules
`BuildOptions` switches on. `build_info` in `schema.sql` says how a finished database records
which ones ran; what follows is only the reasoning that is nowhere in the code.

Rules are applied independently, and only to `title.basics`, where each is a pure predicate
over one record. Nothing anywhere spells `Rated && NotAdult`; that conjunction exists only as
a loop over whichever rules are switched on, so any combination works. This holds for any
rule decidable from `title.basics` - a rule needing a different file would add a pass and an
ordering constraint, which is what makes it expensive rather than the rule itself.

`filtering()` deliberately ignores `People`, so asking for people never makes a build pay for
an allow-list it has no rule to fill. That is the one way the two axes could quietly become
entangled, which is why there is a test for it.

Giving every episode the fate of its parent series, rather than judging it on its own merits,
buys the invariant *every series in the database has all of its episodes* at two costs worth
knowing: a rated episode of an unrated series is dropped, and an adult episode of a clean
series is kept. The alternative punches holes in series, which is worse. "Has a rating" is the
source's own decision rather than a threshold chosen here, which is why the filter has no
tunable knob. It is not a marginal filter: against the 2026-07-27 dump both rules together
refused nearly two titles in five.

Per-rule exclusion counts are deliberately *not* stored: every pass already returns
`counts{read, written}` and logs both, so a filtered pass shows the number where it is useful
without freezing it into the file.

`buildInto` settles the allow-list whole before the first layer writes, so no pass accumulates
it for the next and each stays a function of its file and the filter. Fusing the work into the
titles and episodes passes would save re-reading two files, but it would make the layers
order-dependent in a way transactions-per-layer should not be; the saving is small against a
build measured in tens of minutes.

Each pass checks the filter *after* `read++` and *before* building its row. That ordering is
load-bearing twice over: `counts.read` stays a count of source rows, which is what
`build_sources.rows_read` records as provenance about the file, and the interners never see
values that occur only on refused rows, so a filtered build's lookup tables carry no dead
genres, regions or jobs.

`README.md` carries the remaining work on both axes.

## Tests

Fixtures live uncompressed in `testdata/imdb` so they can be read and diffed; tests that need
the gzipped form a real build sees gzip them into a `t.TempDir()` (see `gzipFixtures` in
`importer/metadata_test.go`). Tests use `testify` (`require` for preconditions, `assert` for
checks) and nested `t.Run` subtests whose names read as sentences, which is what the
`testdox` output in `make test` is for.
