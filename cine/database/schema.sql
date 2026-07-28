-- IMDb non-commercial datasets, the half every build creates.
--
-- The tables divide in two, a file each. A table is Titles if it can be populated
-- without knowing that any person exists; everything else is People, and lives in
-- schema-people.sql, which is applied only when --people asks for it. So a build
-- creates only the tables it is going to fill, and a query reaching for crew in a
-- titles-only database fails on the missing table rather than returning nothing -
-- which tells a database that was never given credits apart from a title that has
-- none. build_info.has_people answers the same question without consulting
-- sqlite_master.
--
-- The division follows the source files exactly - title.basics, title.ratings,
-- title.episode and title.akas against name.basics, title.crew and
-- title.principals - so no import pass straddles it and none ever runs partially.
--
-- Conventions: tconst/nconst identifiers are stored as the integer that remains
-- once the "tt"/"nm" prefix is dropped; IMDb's \N becomes SQL NULL; ratings are
-- stored in tenths (57 means 5.7).
--
-- A singular table name means a lookup: one row per distinct value, always
-- (id, name). Everything else is plural, build_info excepted because a CHECK
-- holds it to one row. Each lookup is prefixed with the table it serves, which
-- leaves it named for the column that references it - akas.region resolves
-- through akas_region, principals.job through principals_job. The two masks
-- break that only by being plural, titles.genres and akas.types holding a set of
-- titles_genre and akas_type ids rather than one.
--
-- Two kinds of reference appear below. Lookup ids that the importer generates
-- itself are declared as foreign keys and always resolve, so a finished build
-- must return nothing from PRAGMA foreign_key_check.
--
-- Columns holding an IMDb identifier - title_id, name_id, episodes.id and
-- episodes.parent_id - carry no foreign key on purpose. The datasets are not
-- closed over their own identifiers: a credit can name a title or person that no
-- file in the download describes, about 10,000 such rows in a recent full build.
-- A multi-day re-download experiment did not clear them, though which ids are
-- orphaned does change from one refresh to the next. They are kept as they are,
-- because they describe the source faithfully. Read these columns with LEFT JOIN;
-- an inner join drops them silently.
--
-- Filtering removes most of them, an id absent from title.basics being absent
-- from the allow-list too, but it is best effort rather than a guarantee: an
-- episode is allowed by its parent, so episodes can hold an id that titles has no
-- row for - 27 of them in the 2026-07-27 dump.

------------------------------------------------------------------------------
-- Build metadata: what this database was made from, and when
------------------------------------------------------------------------------

-- Schema version, readable before any table is trusted to exist.
PRAGMA user_version = 5;

-- One row, written once the build has succeeded.
--
-- The last three columns record the options the build was given, so a query can
-- tell what a database was never given rather than guessing from what it fails to
-- find. The filter_ columns choose which rows are kept; has_people chooses which
-- tables are created at all.
--
-- Each filter rule is applied independently to title.basics. Episodes are not
-- judged by the rules at all - an episode is kept exactly when its parent series
-- is kept, whichever rule decided that - so the rules do not thin out a series.
CREATE TABLE IF NOT EXISTS build_info (
  id                INTEGER PRIMARY KEY CHECK (id = 1),
  cine_version      TEXT    NOT NULL,
  started_at        TEXT    NOT NULL,  -- RFC 3339, UTC
  finished_at       TEXT    NOT NULL,
  filter_rated      INTEGER NOT NULL DEFAULT 0,  -- kept only titles IMDb has rated
  filter_not_adult  INTEGER NOT NULL DEFAULT 0,  -- dropped titles flagged isAdult
  has_people        INTEGER NOT NULL DEFAULT 0   -- imported names, crew and principals
);

-- One row per source file consumed. A missing row means that file was not
-- imported, which is how a table left empty is told apart from one whose file was
-- read: a build without people writes four rows here rather than seven.
--
-- last_modified is the file's own timestamp, which wget copies from the server.
-- The seven files carry timestamps hours apart, so there is no single date for a
-- download and each file has to carry its own.
CREATE TABLE IF NOT EXISTS build_sources (
  file           TEXT PRIMARY KEY,  -- "title.basics.tsv.gz"
  last_modified  TEXT    NOT NULL,  -- RFC 3339, UTC
  bytes          INTEGER NOT NULL,
  rows_read      INTEGER NOT NULL
);

------------------------------------------------------------------------------
-- Titles: titles and their ratings, episodes, alternate titles
------------------------------------------------------------------------------

-- Enumerated titleType values: movie, short, tvEpisode, ...
CREATE TABLE IF NOT EXISTS titles_type (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- One row per genre; id is the bit position used in titles.genres.
--
-- A bitmask rather than a junction, unlike the two name.basics list fields, and
-- the difference is in the data: 0 of the 12,129,448 genre lists in the 2026-07
-- dump are in non-alphabetical order, so IMDb emits a sorted set and the order
-- carries nothing. A mask represents a set exactly, and sorting by name on read
-- recovers IMDb's order, so nothing is lost. It also stays compact where a
-- junction would not - 55% of titles carry a single genre, which as a junction
-- row would pay per-row overhead to store one value.
--
-- The cost is that a mask predicate cannot use an index: genres & (1 << id)
-- always scans titles. That suits genre as a secondary filter alongside an
-- indexed year or rating, and a partial index (... WHERE genres & 512) can cover
-- a hot genre. Should that stop being enough, the junction is derivable at any
-- time by exploding the set bits, because the mask is lossless here.
CREATE TABLE IF NOT EXISTS titles_genre (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- title.basics joined with title.ratings.
CREATE TABLE IF NOT EXISTS titles (
  id               INTEGER PRIMARY KEY,
  title_type       INTEGER NOT NULL REFERENCES titles_type(id),
  primary_title    TEXT    NOT NULL,
  original_title   TEXT,                          -- NULL when equal to primary_title
  is_adult         INTEGER NOT NULL DEFAULT 0,
  start_year       INTEGER,
  end_year         INTEGER,
  runtime_minutes  INTEGER,
  genres           INTEGER NOT NULL DEFAULT 0,    -- bitmask over titles_genre.id
  average_rating   INTEGER,                       -- rating in tenths; NULL if unrated
  num_votes        INTEGER                        -- NULL if unrated
);

-- title.episode: an episode's place in its parent series.
CREATE TABLE IF NOT EXISTS episodes (
  id              INTEGER PRIMARY KEY,
  parent_id       INTEGER NOT NULL,
  season_number   INTEGER,
  episode_number  INTEGER
);

-- Enumerated akas.region: US, GB, ...
CREATE TABLE IF NOT EXISTS akas_region (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- Enumerated akas.language: en, fr, ...
CREATE TABLE IF NOT EXISTS akas_language (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- One row per akas type; id is the bit position used in akas.types.
--
-- A bitmask discards list order, which here is measured rather than assumed: of
-- 19,311,752 type lists in the 2026-07 dump only 429 hold more than one value,
-- and 288 of those are non-alphabetical. So the order is information, but for
-- 0.002% of rows and with no evident meaning - not worth a junction over a
-- 19M-row column. Only 8 distinct values, so the 63-bit ceiling is remote.
CREATE TABLE IF NOT EXISTS akas_type (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- Interned akas.attributes values.
CREATE TABLE IF NOT EXISTS akas_attribute (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- title.akas: localized and alternate titles.
CREATE TABLE IF NOT EXISTS akas (
  title_id           INTEGER NOT NULL,
  ordering           INTEGER NOT NULL,
  title              TEXT    NOT NULL,
  region             INTEGER REFERENCES akas_region(id),
  language           INTEGER REFERENCES akas_language(id),
  types              INTEGER NOT NULL DEFAULT 0,  -- bitmask over akas_type.id
  is_original_title  INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (title_id, ordering)
) WITHOUT ROWID;

-- Many-to-many between akas rows and their attributes.
--
-- Deliberately no position column, unlike the name.basics junctions: of 312,930
-- attribute lists in the 2026-07 dump only 36 hold more than one value, and 23
-- of those are non-alphabetical, so the order lost amounts to 23 rows in 58.6
-- million. Interned rather than a bitmask because there are 164 distinct values,
-- well past what an integer mask holds.
CREATE TABLE IF NOT EXISTS akas_carry_attributes (
  title_id      INTEGER NOT NULL,
  ordering      INTEGER NOT NULL,
  attribute_id  INTEGER NOT NULL REFERENCES akas_attribute(id),
  PRIMARY KEY (title_id, ordering, attribute_id),
  FOREIGN KEY (title_id, ordering) REFERENCES akas(title_id, ordering)
) WITHOUT ROWID;
