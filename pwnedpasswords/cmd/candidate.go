package main

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"errors"

	"pwnedpasswords/database/sqlite"
)

// recordCandidate looks one password up in the cache database and, when it is
// present, upserts it into the passwords table with its breach count.
// The boolean reports whether the password was found.
func recordCandidate(ctx context.Context, write, cache *sqlite.Queries, password string) (bool, error) {
	sum := sha1.Sum([]byte(password))
	count, err := cache.GetHashCount(ctx, sum[:])
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	params := sqlite.UpsertPasswordParams{Password: password, Count: count}
	if err := write.UpsertPassword(ctx, params); err != nil {
		return false, err
	}
	return true, nil
}
