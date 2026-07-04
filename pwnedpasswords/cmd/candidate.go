package main

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"

	"pwnedpasswords/database/sqlite"
	"pwnedpasswords/filter"
)

// checker runs candidate passwords through the membership filter and, on a
// filter hit, looks them up in the hashes table, recording any breach match.
// A nil filter means every candidate is looked up in the hashes table directly.
type checker struct {
	write  *sqlite.Queries
	cache  *sqlite.Queries
	filter *filter.Filter
}

// check processes one candidate: hash it, consult the filter, and on a hit look
// the hash up in the hashes table, upserting any match with its breach count.
// Counts land in t, which the caller folds into the shared progress.
func (c *checker) check(ctx context.Context, t *tally, candidate []byte) error {
	sum := sha1.Sum(candidate)
	if c.filter != nil {
		t.filterQueries++
		if !c.filter.Contains(sum[:]) {
			return nil
		}
	}

	t.hashQueries++
	count, err := c.cache.GetHashCount(ctx, sum[:])
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}

	params := sqlite.UpsertPasswordParams{Password: string(candidate), Count: count}
	if err := c.write.UpsertPassword(ctx, params); err != nil {
		return err
	}
	t.found++
	return nil
}
