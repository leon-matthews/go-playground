-- IMDb non-commercial datasets in a single, additive-layer schema.
--
-- Every build creates all of these tables; the chosen build layer decides which
-- ones get populated, so a query written against a small database keeps working
-- against a larger one. Layer 0 is Core, layer 1 is People, layer 2 is Full.
-- Layer 3 (FTS5) is added by a separate schema.
--
-- Conventions: tconst/nconst identifiers are stored as the integer that remains
-- once the "tt"/"nm" prefix is dropped; IMDb's \N becomes SQL NULL; ratings are
-- stored in tenths (57 means 5.7).
--
-- Two kinds of reference appear below. Lookup ids that the importer generates
-- itself are declared as foreign keys and always resolve, so a finished build
-- must return nothing from PRAGMA foreign_key_check.
--
-- Columns holding an IMDb identifier - title_id, name_id, episodes.id and
-- episodes.parent_id - carry no foreign key on purpose. The datasets themselves
-- are not closed over their own identifiers: a credit can name a title or person
-- that no file in the download describes, about 10,000 such rows in a recent full
-- build. This was first assumed to be a download-timing artefact; a multi-day
-- experiment fetching the files repeatedly disproved that, so the orphans are
-- inherent to the source and no amount of re-downloading resolves them. They are
-- kept as they are, because they do describe the source faithfully. Read these
-- columns with LEFT JOIN; an inner join drops the orphans silently.
--
-- A filtered build is the exception: an id absent from title.basics is never in
-- the allow-list, so filtering removes these orphans along with the rows the
-- rules refused. See build_info.filter_.

------------------------------------------------------------------------------
-- Build metadata: what this database was made from, and when
------------------------------------------------------------------------------

-- Schema version, readable before any table is trusted to exist.
PRAGMA user_version = 2;

-- One row, written once the build has succeeded.
--
-- The filter_ columns record which row filters ran, so a query can tell what a
-- database was never given rather than guessing from what it fails to find.
-- They are a separate axis from layer: layers choose which tables get populated,
-- filters choose which rows.
--
-- Each rule is applied independently to title.basics. Episodes are not judged by
-- the rules at all - an episode is kept exactly when its parent series is kept,
-- whichever rule decided that - so every series here has all of its episodes.
CREATE TABLE IF NOT EXISTS build_info (
  id                INTEGER PRIMARY KEY CHECK (id = 1),
  layer             INTEGER NOT NULL,  -- base layer: 0 Core, 1 People, 2 Full
  cine_version      TEXT    NOT NULL,
  started_at        TEXT    NOT NULL,  -- RFC 3339, UTC
  finished_at       TEXT    NOT NULL,
  filter_rated      INTEGER NOT NULL DEFAULT 0,  -- kept only titles IMDb has rated
  filter_not_adult  INTEGER NOT NULL DEFAULT 0   -- dropped titles flagged isAdult
);

-- One row per source file consumed. A missing row means that file was not
-- imported, which is how an empty table is told apart from an absent one.
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
-- Layer 0 - Core: titles and their ratings
------------------------------------------------------------------------------

-- Enumerated titleType values: movie, short, tvEpisode, ...
CREATE TABLE IF NOT EXISTS titles_types (
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
CREATE TABLE IF NOT EXISTS genres (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- title.basics joined with title.ratings.
CREATE TABLE IF NOT EXISTS titles (
  id               INTEGER PRIMARY KEY,
  title_type       INTEGER NOT NULL REFERENCES titles_types(id),
  primary_title    TEXT    NOT NULL,
  original_title   TEXT,                          -- NULL when equal to primary_title
  is_adult         INTEGER NOT NULL DEFAULT 0,
  start_year       INTEGER,
  end_year         INTEGER,
  runtime_minutes  INTEGER,
  genres           INTEGER NOT NULL DEFAULT 0,    -- bitmask over genres.id
  average_rating   INTEGER,                       -- rating in tenths; NULL if unrated
  num_votes        INTEGER                        -- NULL if unrated
);

------------------------------------------------------------------------------
-- Layer 1 - People: names, episodes, crew
------------------------------------------------------------------------------

-- Enumerated name.basics primaryProfession values: actor, director, ...
CREATE TABLE IF NOT EXISTS professions (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- name.basics.
CREATE TABLE IF NOT EXISTS names (
  id            INTEGER PRIMARY KEY,
  primary_name  TEXT,     -- NULL when the source has \N
  birth_year    INTEGER,
  death_year    INTEGER
);

-- name.basics primaryProfession: the categories a person is credited under.
--
-- A junction rather than a bitmask, because IMDb ranks the list by prominence -
-- Ingmar Bergman is "writer,director,actor" - and 54% of the multi-entry lists
-- in the 2026-07 dump are not in alphabetical order, so the order is content
-- rather than a sorted set. position is 1-based, so position 1 is the primary
-- profession. Carrying the list this way also retires the 63-value ceiling that
-- a single integer mask imposed.
CREATE TABLE IF NOT EXISTS names_primary_professions (
  name_id        INTEGER NOT NULL,
  position       INTEGER NOT NULL,
  profession_id  INTEGER NOT NULL REFERENCES professions(id),
  PRIMARY KEY (name_id, position)
) WITHOUT ROWID;

-- name.basics knownForTitles: the titles a person is best known for.
--
-- position is 1-based and preserves IMDb's own order, matching akas.ordering and
-- principals.ordering. That order carries information rather than being a sorted
-- set: 29% of the lists in the 2026-07 dump are not in tconst order.
CREATE TABLE IF NOT EXISTS names_known_for_titles (
  name_id   INTEGER NOT NULL,
  position  INTEGER NOT NULL,
  title_id  INTEGER NOT NULL,
  PRIMARY KEY (name_id, position)
) WITHOUT ROWID;

-- title.episode: an episode's place in its parent series.
CREATE TABLE IF NOT EXISTS episodes (
  id              INTEGER PRIMARY KEY,
  parent_id       INTEGER NOT NULL,
  season_number   INTEGER,
  episode_number  INTEGER
);

-- title.crew: director and writer credits (role 0 = director, 1 = writer).
--
-- position is 1-based and restarts per title and role, preserving IMDb's order
-- within each list: 2,909,077 of the crew lists in the 2026-07 dump are not in
-- nconst order, so the order is content. It is a payload column rather than part
-- of the key, so the primary key keeps enforcing that nobody is credited twice in
-- the same role on the same title.
--
-- Beware that these lists describe a whole title, so for a long-running series
-- they accumulate decades of crew - up to 1,442 writers and 548 directors on one
-- tvSeries - and position means far less at that length than it does on a film
-- with two directors.
CREATE TABLE IF NOT EXISTS titles_credit_names (
  title_id  INTEGER NOT NULL,
  name_id   INTEGER NOT NULL,
  role      INTEGER NOT NULL,
  position  INTEGER NOT NULL,
  PRIMARY KEY (title_id, name_id, role)
) WITHOUT ROWID;

------------------------------------------------------------------------------
-- Layer 2 - Full: principals and akas
------------------------------------------------------------------------------

-- Enumerated principals.category: actor, director, self, ...
CREATE TABLE IF NOT EXISTS principals_categories (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- Interned free-text principals.job values.
CREATE TABLE IF NOT EXISTS principals_jobs (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- title.principals: one cast or crew credit per row.
CREATE TABLE IF NOT EXISTS principals (
  title_id    INTEGER NOT NULL,
  ordering    INTEGER NOT NULL,
  name_id     INTEGER NOT NULL,
  category    INTEGER NOT NULL REFERENCES principals_categories(id),
  job         INTEGER REFERENCES principals_jobs(id),
  characters  TEXT,                               -- JSON array of names; NULL when absent
  PRIMARY KEY (title_id, ordering),
  CHECK (characters IS NULL OR json_valid(characters))
) WITHOUT ROWID;

-- Enumerated akas.region: US, GB, ...
CREATE TABLE IF NOT EXISTS regions (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- Enumerated akas.language: en, fr, ...
CREATE TABLE IF NOT EXISTS languages (
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
CREATE TABLE IF NOT EXISTS akas_types (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- Interned akas.attributes values.
CREATE TABLE IF NOT EXISTS attributes (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- title.akas: localized and alternate titles.
CREATE TABLE IF NOT EXISTS akas (
  title_id           INTEGER NOT NULL,
  ordering           INTEGER NOT NULL,
  title              TEXT    NOT NULL,
  region             INTEGER REFERENCES regions(id),
  language           INTEGER REFERENCES languages(id),
  types              INTEGER NOT NULL DEFAULT 0,  -- bitmask over akas_types.id
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
  attribute_id  INTEGER NOT NULL REFERENCES attributes(id),
  PRIMARY KEY (title_id, ordering, attribute_id),
  FOREIGN KEY (title_id, ordering) REFERENCES akas(title_id, ordering)
) WITHOUT ROWID;
