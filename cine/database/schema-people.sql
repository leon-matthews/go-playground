-- The People half of the schema, applied only when --people asks for it.
--
-- schema.sql carries the conventions these tables follow, and the reasoning
-- behind the division. Nothing here is referenced from there, which is what lets
-- the file be left out: the lookups below are the only foreign key parents, and
-- title_id columns carry no foreign key at all.

------------------------------------------------------------------------------
-- People: names, crew and principals
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
--
-- A filtered build leaves a gap where it refused a title: unfiltered lists are
-- always 1..n, so renumbering would assert a rank IMDb never published.
CREATE TABLE IF NOT EXISTS names_known_for_titles (
  name_id   INTEGER NOT NULL,
  position  INTEGER NOT NULL,
  title_id  INTEGER NOT NULL,
  PRIMARY KEY (name_id, position)
) WITHOUT ROWID;

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
