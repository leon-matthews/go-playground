
-- Hashes are fetched by their 5-character prefix
CREATE TABLE prefixes (
  id      INTEGER PRIMARY KEY,
  prefix  TEXT    NOT NULL UNIQUE,
  updated INTEGER,
  etag    TEXT,
  hashes  TEXT NOT NULL -- SHA1 of passwords and their counts
);
