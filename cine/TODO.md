# TODO

## Filters

Building a much smaller database by keeping only titles popular enough that somebody
bothered to rate them. The faithful full build stays as it is; this is a second, smaller
artefact. All numbers below are measured against the 2026-07-26 dump in `~/Temp/imdb` and
the 6.45 GiB full build (`cine.full.db`), not estimated.

This is a separate axis from the `layer` idea in `database/schema.sql`. Layers choose which
*tables* get populated; this chooses which *rows*. Layers remain an unimplemented idea under
consideration.

### The base filter

IMDb only publishes a rating once a title has 5 votes, so "has a rating at all" is the
source's own popularity decision rather than a number we invented. That alone takes
12,669,117 titles down to 1,698,599 - 13.4%.

| titleType | all | rated | rated % | adult (all) |
|---|---:|---:|---:|---:|
| tvEpisode | 9,791,652 | 871,787 | 8.9% | 279,084 |
| short | 1,146,209 | 185,735 | 16.2% | 2,739 |
| movie | 753,019 | 348,355 | 46.3% | 9,121 |
| video | 328,643 | 59,675 | 18.2% | 115,535 |
| tvSeries | 302,817 | 113,388 | 37.4% | 3,122 |
| tvMovie | 155,508 | 57,022 | 36.7% | 87 |
| tvMiniSeries | 71,778 | 25,747 | 35.9% | 477 |
| tvSpecial | 58,771 | 14,268 | 24.3% | 36 |
| videoGame | 49,686 | 20,017 | 40.3% | 2,483 |
| tvShort | 11,033 | 2,605 | 23.6% | 3 |
| tvPilot | 1 | 0 | 0.0% | 0 |
| **TOTAL** | **12,669,117** | **1,698,599** | **13.4%** | **412,687** |

Adult titles are 412,687 overall but only 25,513 rated, so excluding them is a content
decision worth 1.5% of the small database - not a size lever.

`tvPilot` has exactly one row and it is unrated, so that type will silently vanish from
`titles_types` in any filtered build. Worth being deliberate about.

### Episodes are the whole problem

Episodes are 77% of all titles and still 51% of rated ones. The requirement is that a
series must never be stored with holes in it: if a series is included, every episode of it
is included.

- 239,305 distinct parent series in `title.episode`, holding 9,791,652 episodes
- 122,998 of those parents are rated (51.4%), holding 7,275,393 episodes
- Episodes themselves are rated at only 8.9%, so parents rate ~6x higher than their leaves
- Of 871,787 rated episodes, 870,292 (99.83%) already have a rated parent - only 1,495
  rated episodes hang off an unrated series

Rated series broken down by how much of each is covered by episode ratings:

| coverage of the series | series | their episodes | rated episodes |
|---|---:|---:|---:|
| none of it | 76,621 | 4,250,767 | 0 |
| 1-9% | 7,019 | 1,840,898 | 25,120 |
| 10-49% | 7,621 | 365,521 | 87,573 |
| 50-89% | 5,612 | 191,063 | 135,840 |
| 90-99% | 1,921 | 152,466 | 147,081 |
| all of it | 24,204 | 474,678 | 474,678 |
| **total** | **122,998** | **7,275,393** | **870,292** |

### Rejected: episode coverage is not an artefact of collection history

Hypothesis was that IMDb only recently started collecting episode-level ratings, which
would make sparse coverage a proxy for age rather than for obscurity. Measured coverage by
episode air decade and it is flat - between 8.6% and 10.6% for seventy years, with no drift
toward the present:

| decade | episodes | rated | % |
|---|---:|---:|---:|
| pre-1950 | 6,073 | 132 | 2.2% |
| 1950s | 133,407 | 12,935 | 9.7% |
| 1960s | 289,447 | 26,670 | 9.2% |
| 1970s | 360,242 | 31,001 | 8.6% |
| 1980s | 392,448 | 39,410 | 10.0% |
| 1990s | 632,933 | 67,124 | 10.6% |
| 2000s | 1,277,645 | 134,779 | 10.5% |
| 2010s | 3,002,003 | 302,758 | 10.1% |
| 2020s | 2,417,434 | 256,795 | 10.6% |
| **unknown** | **1,280,020** | **183** | **0.0%** |

What coverage does track is popularity. Median vote count on the series itself, by band:

| coverage | series | median votes | 90th pct | mean episodes |
|---|---:|---:|---:|---:|
| 0% | 76,621 | 21 | 98 | 55 |
| 1-49% | 14,640 | 114 | 592 | 151 |
| 50-99% | 7,533 | 247 | 3,380 | 46 |
| 100% | 24,204 | 672 | 10,019 | 20 |

Incidental find: 1,280,020 episodes carry no `startYear` at all and 183 of them are rated.

### Rejected: rule B, "keep a series' episodes if any one of them is rated"

Superficially attractive - one sentence, no threshold, hole-free, and 3,872,628 titles. But
the threshold is one vote, which is the sharpest cliff available, and it splits similar
shows arbitrarily:

- 6,527 series pull in 598,750 episodes on the strength of a single rated episode
- 6,139 *excluded* series have more votes than the median *included* series, and hold
  787,629 episodes between them
- 7,280 included series have fewer votes than that same median

Concrete case: Nintendo Direct (46 votes, 2 of 121 episodes rated) is kept whole, while
Shanti (109 votes, 778 episodes, none rated) is dropped entirely. Twice the votes, excluded.

A percentage threshold (rule C, "at least half the episodes rated") fails for a related
reason: recent shows have partial coverage simply because votes are still accumulating.
Dark Matter is 9/19 and Dune: Prophecy is 6/14, so a 50% bar punches holes in exactly the
new prestige TV people search for, and does it differently on every rebuild.

### Current front-runner: rule A

> Keep a title if it is rated, or if it is an episode of a rated series.

Keys on a property of the series itself, so it is stable across rebuilds and independent of
episode-level accidents. The trailing clause keeps the filter monotone - nothing that
carries a rating is ever discarded, including the 1,495 rated episodes whose parent is
unrated. Yields a keep-set of 8,103,700 titles (64.0%).

It also buys an invariant worth writing into `database/schema.sql`: *every series in this
database has all of its episodes*.

### Measured cascade

Retention of each source file under each candidate, measured by streaming the TSVs against
each keep-set:

| | principals | akas | crew/titles | episodes |
|---|---:|---:|---:|---:|
| rated only | 23.2% | 13.5% | 13.4% | 8.9% |
| A | 73.5% | 70.8% | 64.0% | 74.3% |
| B | 42.2% | 31.7% | 30.6% | 31.1% |
| C | 23.7% | 13.7% | 13.9% | 9.5% |

Full build table sizes from `dbstat`, and the projected result of each filter (MiB):

| | principals | akas | titles | episodes | credits | people¹ | **total** | vs full |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| full | 2,521 | 2,106 | 546 | 155 | 355 | 913 | **6,600** | 100% |
| **A** | 1,853 | 1,491 | 349 | 115 | 227 | 913 | **4,951** | **75%** |
| B | 1,064 | 668 | 167 | 48 | 109 | 913 | **2,972** | 45% |
| C | 597 | 288 | 76 | 15 | 49 | 913 | **1,941** | 29% |
| rated only | 585 | 284 | 73 | 14 | 48 | 913 | **1,920** | 29% |

¹ `names` + `names_known_for_titles` + `names_primary_professions`, which no title filter
touches.

**Rule A takes 6.45 GiB to roughly 4.8 GiB - a 25% reduction.** Episodes are not cheap:
they carry 73.5% of `principals` and 70.8% of `akas`, because every episode gets its own
cast list and its own set of localised titles. So "no partial series" and "much smaller
database" are in genuine conflict, and one of them has to give.

### Open questions

1. **Does 4.8 GiB count as success?** If not, the invariant has to be bought some other way
   than by lowering the bar to IMDb's 5-vote floor.
2. **A vote threshold on the series keeps the invariant and cuts size.** "Keep a series if
   it has at least N votes, then keep every one of its episodes" has the same one-sentence
   shape as rule A and the same no-holes guarantee. Cascade not yet measured.

   | series bar | series kept | their episodes |
   |---|---:|---:|
   | any rating (= A) | 122,998 | 7,275,393 |
   | >= 50 votes | 62,038 | 5,145,486 |
   | >= 100 votes | 47,152 | 4,411,214 |
   | >= 500 votes | 22,564 | 2,724,780 |
   | >= 1,000 votes | 15,943 | 2,120,588 |

3. **The people tables are untouched by any title filter** - 913 MiB, which under the
   smaller filters would be nearly half the file. Restricting `names` to people actually
   credited in a kept title is an independent lever, and under rule C it is worth more than
   the choice of title filter itself. Not yet measured.
4. **Where does the filter live?** `readRatings` in `importer/importer.go:222` already loads
   the ratings map before any layer runs, so the keep-set can be built there. Episodes need
   a pre-pass over `title.episode` to learn parentage before `title.basics` is streamed.
5. **How is a filtered build recorded?** `build_info` has a `layer` column but nothing that
   says which row filter produced the file. A query written against a filtered database
   should be able to tell what was left out.
6. **Adult titles** - excluded or not? Worth 1.5% of the small database, so a content call
   rather than a size one.
