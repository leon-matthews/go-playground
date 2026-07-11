// Package checker runs candidate passwords through the membership filter and
// the breach-hash table, recording any match in the output database.
package checker

import (
	"bytes"
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"

	"pwnedpasswords/database"
	"pwnedpasswords/database/sqlite"
	"pwnedpasswords/filter"
	"pwnedpasswords/progress"
)

// Checker runs candidate passwords through the membership filter and, on a
// filter hit, looks them up in the hashes table, recording any breach match.
// A nil Filter means every candidate is looked up in the hashes table directly.
type Checker struct {
	Writer *database.BatchWriter
	Cache  *sqlite.Queries
	Filter *filter.Filter
}

// Check processes one candidate: hash it, consult the filter, and on a hit look
// the hash up in the hashes table, upserting any match with its breach count.
// Counts land in t, which the caller folds into the shared progress.
func (c *Checker) Check(ctx context.Context, t *progress.Tally, candidate []byte) error {
	sum := sha1.Sum(candidate)
	if c.Filter != nil {
		t.FilterQueries++
		if !c.Filter.Contains(sum) {
			return nil
		}
	}

	t.HashQueries++
	// Clone only on the rare filter hit; passing sum[:] would escape sum every call.
	count, err := c.Cache.GetHashCount(ctx, bytes.Clone(sum[:]))
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	password := string(candidate)
	changed, err := c.Writer.Upsert(ctx, password, count)
	if err != nil {
		return err
	}
	t.Found++
	t.Sample = password
	if changed > 0 {
		t.Changed++
	}
	return nil
}
