-- Passwords found in the breach corpus, with the number of breaches each appeared in.
CREATE TABLE IF NOT EXISTS passwords (
  password TEXT PRIMARY KEY,
  count    INTEGER NOT NULL
);

-- Serve top-N denylist exports as an ordered index scan, without a sort.
CREATE INDEX IF NOT EXISTS passwords_by_count ON passwords (count DESC);
