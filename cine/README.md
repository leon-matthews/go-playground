# Cine

A tool to build and query a local SQLite3 database containing data from the
IMDb Non-Commercial Datasets, that are licensed for personal and non-commercial use.

Download all of them in one shot using `wget`. The `-N` flag re-fetches only those files
whose timestamp or size has changed, so re-running it picks up IMDb's daily refresh:

    $ wget -A gz -r -N -l 1 -nd -e robots=off https://datasets.imdbws.com/

## TODO

Now that filtering has been implemented, let's look into a layering concept.

My first idea is that we start with describing all the titles in the base layer,
then, if the user wishes, add another layer on top capturing the people in those
titles:

The two-layer split, sized (fresh dbstat on the new build):

┌──────────┬────────────────────────────────────────────────────────────────────────────────────────────────────────────┬───────┐
│  layer   │                                                   tables                                                   │  MiB  │
├──────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────┼───────┤
│ 1 Titles │ titles 357.9, akas 1671.4, episodes 117.7, akas_carry_attributes 2.6, lookups                              │ 2,150 │
├──────────┼────────────────────────────────────────────────────────────────────────────────────────────────────────────┼───────┤
│ 2 People │ principals 2136.2, names 393.8, titles_credit_names 291.7, known-for 245.2, professions 211.9, job lookups │ 3,282 │
└──────────┴────────────────────────────────────────────────────────────────────────────────────────────────────────────┴───────┘

40/60, totalling the 5.30 GiB on disk. And the split has a rule that sorts
every table without argument: a table is Titles if you can populate it without
knowing that any person exists. episodes moves up from People - an episode is
a title, and that table has no person in it; it was only in People
because of how the import was batched. akas moves down from Full.

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
