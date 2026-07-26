# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this
repository.

`cine` builds a local SQLite database from the IMDb non-commercial datasets - seven gzipped
TSV files fetched with the `wget` line in `README.md`. A full build is ~6.5 GiB and takes
tens of minutes, so most work is done against the small TSV fixtures in `testdata/imdb`.

## Commands

Run everything from this folder (`cine/`), which holds `go.mod`.

- `make test` - `gotestsum` with testdox output and `--failfast`
- `make lint` - `revive ./...`
- `go vet ./...` then `revive ./...` then `gofumpt --extra -w .` to validate changes
- One test: `go test ./importer -run TestBuildMetadata` (subtests via `-run 'Test/sub'`)
- Benchmark: `go test ./importer -run '^$' -bench BenchmarkInsertChunkSizes`
- Fuzz: `go test ./reader -run '^$' -fuzz FuzzParseCharacters`
- Regenerate sqlc code: `sqlc generate` from `database/` (config paths are relative to it)
- Real build: `./cine build-database ~/Temp/imdb cine.db`; `--profile` writes `cpu.pprof`
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
  "layer" imported in its own transaction, in dependency order, from the slice in
  `buildInto`. `title.ratings` has no table: it is loaded into a map first and joined onto
  titles during their pass, and its only record is its `build_sources` row.
- **`database`** embeds `schema.sql`, applies it on `Open`, and tunes the connection for
  bulk loading (no journal, no fsync, exclusive lock, one connection). `Open` is
  write-focused; a read path should open the finished file separately with gentler pragmas.
- **`cmd`** is cobra wiring only - argument checking, the logger, the `--profile` flag.

### Conventions worth knowing before editing

- **`schema.sql` is the design document.** Every non-obvious column choice is justified in a
  comment there with measured counts from a real dump. Read it before changing a table.
- **Bitmask vs junction is a measured decision, not a preference.** A list field becomes a
  bitmask when its order carries nothing (genres, akas types) and a junction with a 1-based
  `position` when IMDb's order is content (primary professions, known-for titles, crew).
  The `interner` serves both: `id()` for lookup foreign keys, `bit()` for bitmask positions.
- **Columns holding IMDb identifiers carry no foreign key on purpose.** The datasets are not
  closed over their own identifiers - credits reference titles and people no file in the
  download describes. This was assumed to be a download-timing artefact until a multi-day
  experiment disproved it, so the orphans are inherent to the source and re-downloading does
  not resolve them. Read these columns with `LEFT JOIN`. Generated lookup ids *are* foreign
  keys and must always resolve - a finished build returns nothing from
  `PRAGMA foreign_key_check`.
- **Bulk inserts are hand-written, not sqlc.** sqlc cannot emit multi-row SQLite inserts.
  `batchInserter` chunks rows to `bindParamBudget` (36) bind parameters, a figure measured by
  `BenchmarkInsertChunkSizes` against the modernc driver - don't raise it on intuition.
  `database/sqlite/*.go` is generated; edit `query.sql` and regenerate instead.
- **Each import pass follows the same shape**: `xColumns` slice, a `bindXRow` function that
  appends values in that order, a `buildXRow` transformer, and an `importX` streaming loop
  returning `counts{read, written}` plus any interners for the caller to flush.
- Comments are one line where possible, and API docs put the first sentence on its own line.

### The row filter

`importer/filter.go` builds the allow-list for a smaller second database. `FilterRules`
selects which rules run, and there are two so far - `Rated` and `NotAdult`. The zero value
filters nothing, which is the full build.

Rules are applied independently, and only to `title.basics`, where both are pure predicates
over one record. Nothing anywhere spells `Rated && NotAdult`; that conjunction exists only as
a loop over whichever rules are switched on, so any combination works. This holds for any
rule decidable from `title.basics` - a rule needing a different file would add a pass and an
ordering constraint, which is what makes it expensive rather than the rule itself.

Episodes are never judged by the rules:

> An episode is kept exactly when its parent series is kept, whichever rule decided that.

So `readEpisodes` never asks *why* a parent was allowed, and that is deliberate. It buys the
invariant *every series in the database has all of its episodes*, at two costs worth knowing:
a rated episode of an unrated series is dropped, and an adult episode of a clean series is
kept. The alternative - judging episodes individually - punches holes in series, which is
worse. "Has a rating" is the source's own decision rather than a threshold chosen here, which
is why the filter has no tunable knob. Measured against the 2026-07 dump, the rated rule
leaves roughly 64% of titles, taking a 6.45 GiB full build to about 4.8 GiB.

`build_info.filter_rated` and `filter_not_adult` record which rules ran. The same
`FilterRules` value drives the filter and is written to those columns, so a database cannot
claim a rule that did not run. Per-rule exclusion counts are deliberately *not* stored: every
pass already returns `counts{read, written}` and logs both, so a filtered pass shows the
number where it is useful without freezing it into the file.

`build-database` takes `--allow-adult` and `--allow-unrated`, both off, so the default run
asks for both rules. The flags name what to keep and `FilterRules` names what to restrict,
so `cmd` inverts each once.

`buildInto` calls `buildFilter` just after the ratings load - where `ratings` exists and
still before any layer writes - and hands the result to the five title-keyed passes: titles,
episodes, crew, principals and akas, but not names. The allow-list is settled whole before
the first write, so no pass accumulates it for the next and each stays a function of its file
and the filter. Fusing the work into the titles and episodes passes would save re-reading two
files, but it would make the layers order-dependent in a way transactions-per-layer should
not be; the saving is small against a build measured in tens of minutes.

Each pass checks the filter *after* `read++` and *before* building its row. That ordering is
load-bearing twice over: `counts.read` stays a count of source rows, which is what
`build_sources.rows_read` records as provenance about the file, and the interners never see
values that occur only on refused rows, so a filtered build's lookup tables carry no dead
genres, regions or jobs. `importEpisodes` checks the episode id alone - the filter allows an
episode only where it kept the parent, so a second check would be redundant.

**A filtered build also drops the source's inherent orphans**, because an id absent from
`title.basics` is never in the allow-list. This is deliberate. Unfiltered, those rows are kept
as a faithful record of the source; filtered, no reader can tell an orphan from a title the
rules refused, and rows nothing can join to are only bytes.

Remaining work:

- Whether ~4.8 GiB is small enough. If not, a vote threshold on the series would keep the
  invariant and cut deeply, but it reintroduces a magic number the current rules avoid.
- The `names` tables are ~913 MiB and no title filter touches them. Restricting them to
  people credited in a kept title is an independent, larger lever - but the ids needed to
  decide that only appear once crew and principals have streamed, which is after `names`
  imports, so it forces a change to the layer ordering.
- `names_known_for_titles` rows can now point at titles the filter dropped, since the names
  pass is unfiltered. Deciding that needs the same title allow-list the other passes take,
  so it is cheap to do and only waiting on whether those rows are wanted.

The `layer` column in `build_info` is a separate, unimplemented axis: layers would choose
which *tables* get populated, the filter chooses which *rows*. Only `layerFull` (2) is
produced today.

## Tests

Fixtures live uncompressed in `testdata/imdb` so they can be read and diffed; tests that need
the gzipped form a real build sees gzip them into a `t.TempDir()` (see `gzipFixtures` in
`importer/metadata_test.go`). Tests use `testify` (`require` for preconditions, `assert` for
checks) and nested `t.Run` subtests whose names read as sentences, which is what the
`testdox` output in `make test` is for.
