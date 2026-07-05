-- Read-only source table living in the sibling pwnedcache database.
-- Declared here only so sqlc can type queries against it; never created by this tool.
CREATE TABLE hashes (
  hash  BLOB PRIMARY KEY,
  count INTEGER NOT NULL
) WITHOUT ROWID;
