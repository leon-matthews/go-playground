-- Passwords found in the breach corpus, with the number of breaches each appeared in.
-- WITHOUT ROWID makes the password the physical key: one B-tree, stored once, single
-- descent per select/insert/update. No secondary index, so writes touch only this tree.
CREATE TABLE IF NOT EXISTS passwords (
  password TEXT PRIMARY KEY,
  count    INTEGER NOT NULL
) WITHOUT ROWID;
