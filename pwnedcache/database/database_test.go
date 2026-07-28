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

	"pwnedcache/database"
	"pwnedcache/database/sqlite"
)

// open builds a fresh database in a per-test temporary folder.
func open(t testing.TB) (*sqlite.Queries, *sql.DB) {
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

func TestPragmas(t *testing.T) {
	ctx := context.Background()
	_, db := open(t)

	// Each tuning pragma should be live on the connection Open handed back
	cases := []struct {
		pragma string
		want   string
	}{
		{"journal_mode", "wal"},
		{"locking_mode", "exclusive"},
		{"synchronous", "1"}, // NORMAL
		{"cache_size", "-65536"},
		{"temp_store", "2"}, // MEMORY
	}
	for _, tc := range cases {
		t.Run(tc.pragma, func(t *testing.T) {
			var got string
			err := db.QueryRowContext(ctx, "PRAGMA "+tc.pragma).Scan(&got)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestOpenReadOnly(t *testing.T) {
	ctx := context.Background()

	t.Run("missing database is not created", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing.db")
		_, _, err := database.OpenReadOnly(ctx, path)
		assert.ErrorIs(t, err, database.ErrNotFound)
		assert.NoFileExists(t, path)
	})

	t.Run("existing database is queryable", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.db")
		queries, db, err := database.Open(ctx, path)
		require.NoError(t, err)
		hash := fromHex(t, "cafe5"+strings.Repeat("0", 35))
		require.NoError(t, queries.InsertHash(ctx, sqlite.InsertHashParams{Hash: hash, Count: 7}))
		require.NoError(t, db.Close())

		queries, db, err = database.OpenReadOnly(ctx, path)
		require.NoError(t, err)
		defer db.Close()

		count, err := queries.GetHashCount(ctx, hash)
		require.NoError(t, err)
		assert.Equal(t, int64(7), count)

		// Writes are refused even though the statements prepare cleanly
		row := sqlite.InsertHashParams{Hash: fromHex(t, strings.Repeat("a", 40)), Count: 1}
		assert.Error(t, queries.InsertHash(ctx, row))
	})

	// Each of these would be read as URI syntax if the path were not escaped
	t.Run("path needing escaping", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "a b?c#d.db")
		_, db, err := database.Open(ctx, path)
		require.NoError(t, err)
		require.NoError(t, db.Close())

		// Open wrote to the exact path asked for, not a truncated one
		assert.FileExists(t, path)

		_, db, err = database.OpenReadOnly(ctx, path)
		require.NoError(t, err)
		defer db.Close()
		assert.NoError(t, db.PingContext(ctx))
	})
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
