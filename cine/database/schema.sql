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

------------------------------------------------------------------------------
-- Layer 0 - Core: titles and their ratings
------------------------------------------------------------------------------

-- Enumerated titleType values: movie, short, tvEpisode, ...
CREATE TABLE IF NOT EXISTS titles_types (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- One row per genre; id is the bit position used in titles.genres.
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

-- One row per profession; id is the bit position used in names.primary_profession.
CREATE TABLE IF NOT EXISTS professions (
  id    INTEGER PRIMARY KEY,
  name  TEXT    NOT NULL UNIQUE
);

-- name.basics.
CREATE TABLE IF NOT EXISTS names (
  id                  INTEGER PRIMARY KEY,
  primary_name        TEXT    NOT NULL,
  birth_year          INTEGER,
  death_year          INTEGER,
  primary_profession  INTEGER NOT NULL DEFAULT 0  -- bitmask over professions.id
);

-- title.episode: an episode's place in its parent series.
CREATE TABLE IF NOT EXISTS episodes (
  id              INTEGER PRIMARY KEY REFERENCES titles(id),
  parent_id       INTEGER NOT NULL REFERENCES titles(id),
  season_number   INTEGER,
  episode_number  INTEGER
);

-- title.crew: director and writer credits (role 0 = director, 1 = writer).
CREATE TABLE IF NOT EXISTS titles_credit_names (
  title_id  INTEGER NOT NULL REFERENCES titles(id),
  name_id   INTEGER NOT NULL REFERENCES names(id),
  role      INTEGER NOT NULL,
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
  title_id    INTEGER NOT NULL REFERENCES titles(id),
  ordering    INTEGER NOT NULL,
  name_id     INTEGER NOT NULL REFERENCES names(id),
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
  title_id           INTEGER NOT NULL REFERENCES titles(id),
  ordering           INTEGER NOT NULL,
  title              TEXT    NOT NULL,
  region             INTEGER REFERENCES regions(id),
  language           INTEGER REFERENCES languages(id),
  types              INTEGER NOT NULL DEFAULT 0,  -- bitmask over akas_types.id
  is_original_title  INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY (title_id, ordering)
) WITHOUT ROWID;

-- Many-to-many between akas rows and their attributes.
CREATE TABLE IF NOT EXISTS akas_carry_attributes (
  title_id      INTEGER NOT NULL,
  ordering      INTEGER NOT NULL,
  attribute_id  INTEGER NOT NULL REFERENCES attributes(id),
  PRIMARY KEY (title_id, ordering, attribute_id),
  FOREIGN KEY (title_id, ordering) REFERENCES akas(title_id, ordering)
) WITHOUT ROWID;
