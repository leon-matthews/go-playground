
-- Download metadata for each 5-character prefix
CREATE TABLE IF NOT EXISTS prefixes (
  id      INTEGER PRIMARY KEY,
  prefix  TEXT    NOT NULL UNIQUE,
  updated INTEGER,
  etag    TEXT
);

-- Full 20-byte SHA-1 of each password, and its breach count
CREATE TABLE IF NOT EXISTS hashes (
  hash  BLOB PRIMARY KEY,
  count INTEGER NOT NULL
) WITHOUT ROWID;
