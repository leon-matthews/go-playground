# Cine

A tool to build and query a local SQLite3 database containing data from the
IMDb Non-Commercial Datasets, that are licensed for personal and non-commercial use.

Download all of them in one shot using `wget`. The `-N` flag re-fetches only those files
whose timestamp or size has changed, so re-running it picks up IMDb's daily refresh:

    $ wget -A gz -r -N -l 1 -nd -e robots=off https://datasets.imdbws.com/

## Building a database

    $ cine build-database ~/Temp/imdb cine.db

The database is built in a temporary file and renamed into place only once the whole
import succeeds, so an interrupted run leaves any database already there untouched.

Two independent choices decide how much of IMDb ends up in the file. Both default to
the smaller database, and all three flags are recorded in the `build_info` table, so a
query can tell what a database was never given rather than guessing from what it fails
to find.

**Which rows.** By default a title is kept only if IMDb has published a rating for it
and has not flagged it as adult. `--allow-unrated` and `--allow-adult` each switch one
of those rules off, and passing both imports every row. Episodes are never judged by
the rules: an episode is kept exactly when its parent series is kept, so a series is
never stored with gaps in it.

**Which tables.** By default only the titles are imported. `--people` adds
`name.basics`, `title.crew` and `title.principals` as well. The division is that a
table belongs to Titles if it can be populated without knowing that any person exists,
which sorts every table without argument and follows the seven source files exactly, so
no import pass straddles it:

| Group  | Source files                                                    |
|--------|-----------------------------------------------------------------|
| Titles | `title.basics`, `title.ratings`, `title.episode`, `title.akas`  |
| People | `name.basics`, `title.crew`, `title.principals`                 |

A build creates only the tables it is going to fill, so without `--people` the people
tables are absent rather than empty and a query reaching for them fails outright. That
is deliberate: an empty result would answer "this person has no credits" when the truth
is "this database has no credits".

People is much the larger of the two, `principals` alone outweighing every Titles table
put together. Ask `dbstat` for the breakdown of a database you have actually built.

## TODO

- No query side yet, despite the description above: `build-database` and
  `reader-benchmark` are the only two sub-commands. `database.Open` is tuned for bulk
  loading, so a read path wants to open the finished file with gentler pragmas.
- Full-text search over titles. It was going to be a third layer, but it indexes rows
  that are already there rather than adding source data, so it belongs beside
  `has_people` as its own column and flag.
- Whether the row filters cut deeply enough. A vote threshold on the series would keep
  the every-episode invariant and cut much harder, but it reintroduces a magic number
  that "has a rating" avoids by deferring to the source's own decision.
- Restricting `names` to people credited in a kept title. Now a much weaker lever than
  it was, since anyone who wants a small database can leave `--people` off entirely, and
  it needs ids that only exist once crew and principals have streamed - which is after
  `names` imports, so it would force the passes back into a dependency order.

## Copyright

Copyright 2026 Leon Matthews. All rights reserved.

## License

This project is licensed under the Apache License 2.0. You are free to use, modify, and
distribute this software for personal or commercial purposes. In return, you must:

1. Include a copy of the license in any redistribution.
2. Retain all copyright, patent, trademark, and attribution notices.
3. State any significant changes you make to the files.

The software is provided "as is", without warranty of any kind. This summary
is for convenience only; the LICENSE file is the authoritative text.
