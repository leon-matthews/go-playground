package database_test

import (
	"context"
	"database/sql"
	"encoding/hex"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pwneddb/database"
	"pwneddb/database/sqlite"
)

// open builds a fresh database in a per-test temporary folder.
func open(t *testing.T) (*sqlite.Queries, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "test.db")
	queries, db, err := database.Open(ctx, path)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return queries, db
}

// fromHex decodes a hexadecimal string, failing the test on error.
func fromHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func TestOpenAndUpsert(t *testing.T) {
	ctx := context.Background()
	queries, _ := open(t)

	insert := sqlite.UpsertPrefixParams{
		Prefix:  "cafe5",
		Updated: sql.NullInt64{Int64: 100, Valid: true},
		Etag:    sql.NullString{String: "etag-1", Valid: true},
	}
	require.NoError(t, queries.UpsertPrefix(ctx, insert))

	row, err := queries.GetPrefix(ctx, "cafe5")
	require.NoError(t, err)
	assert.Equal(t, "etag-1", row.Etag.String)
	assert.Equal(t, int64(100), row.Updated.Int64)

	// Upserting the same prefix updates the row in place
	update := sqlite.UpsertPrefixParams{
		Prefix:  "cafe5",
		Updated: sql.NullInt64{Int64: 200, Valid: true},
		Etag:    sql.NullString{String: "etag-2", Valid: true},
	}
	require.NoError(t, queries.UpsertPrefix(ctx, update))

	row, err = queries.GetPrefix(ctx, "cafe5")
	require.NoError(t, err)
	assert.Equal(t, "etag-2", row.Etag.String)
	assert.Equal(t, int64(200), row.Updated.Int64)

	// The conflict updated rather than inserted, so still one row
	etags, err := queries.GetEtags(ctx)
	require.NoError(t, err)
	assert.Len(t, etags, 1)
}

func TestHashQueries(t *testing.T) {
	ctx := context.Background()
	queries, _ := open(t)

	// Two hashes inside the "cafe5" prefix range, one outside it
	inside1 := fromHex(t, "cafe5"+"003d68eb55068c33ace09247ee4c639306b")
	inside2 := fromHex(t, "cafe5"+strings.Repeat("f", 35))
	outside := fromHex(t, "cafe6"+strings.Repeat("0", 35))
	for i, hash := range [][]byte{inside1, inside2, outside} {
		row := sqlite.InsertHashParams{Hash: hash, Count: int64(i + 1)}
		require.NoError(t, queries.InsertHash(ctx, row))
	}

	count, err := queries.GetHashCount(ctx, inside1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Range delete removes the whole prefix, including its upper bound
	bounds := sqlite.DeleteHashRangeParams{
		Lower: fromHex(t, "cafe5"+strings.Repeat("0", 35)),
		Upper: fromHex(t, "cafe5"+strings.Repeat("f", 35)),
	}
	require.NoError(t, queries.DeleteHashRange(ctx, bounds))

	_, err = queries.GetHashCount(ctx, inside1)
	assert.ErrorIs(t, err, sql.ErrNoRows)
	_, err = queries.GetHashCount(ctx, inside2)
	assert.ErrorIs(t, err, sql.ErrNoRows)

	// The neighbouring prefix is untouched
	count, err = queries.GetHashCount(ctx, outside)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}
