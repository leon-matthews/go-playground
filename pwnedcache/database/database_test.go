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

// pragma reads one pragma's value from the given connection.
func pragma(t *testing.T, db *sql.DB, name string) string {
	t.Helper()
	var got string
	require.NoError(t, db.QueryRowContext(context.Background(), "PRAGMA "+name).Scan(&got))
	return got
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

	// assertPragmas checks each pragma is live on the connection under test
	assertPragmas := func(t *testing.T, db *sql.DB, cases [][2]string) {
		for _, tc := range cases {
			t.Run(tc[0], func(t *testing.T) {
				assert.Equal(t, tc[1], pragma(t, db, tc[0]))
			})
		}
	}

	t.Run("bulk load", func(t *testing.T) {
		_, db := open(t)
		assertPragmas(t, db, [][2]string{
			{"journal_mode", "wal"},
			{"locking_mode", "exclusive"},
			{"synchronous", "1"}, // NORMAL
			{"cache_size", "-65536"},
			{"temp_store", "2"}, // MEMORY
		})
	})

	t.Run("read only", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.db")
		_, db, err := database.Open(ctx, path)
		require.NoError(t, err)
		require.NoError(t, database.Close(ctx, db))

		_, db, err = database.OpenReadOnly(ctx, path)
		require.NoError(t, err)
		defer db.Close()

		assertPragmas(t, db, [][2]string{
			{"query_only", "1"},
			{"cache_size", "-65536"},
			{"temp_store", "2"}, // MEMORY
			// Opening to read leaves the journal mode Close settled on
			{"journal_mode", "delete"},
			{"locking_mode", "normal"},
		})
	})
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

func TestClose(t *testing.T) {
	ctx := context.Background()

	// journalMode reports the mode recorded in the database file itself
	journalMode := func(t *testing.T, path string) string {
		t.Helper()
		_, db, err := database.OpenReadOnly(ctx, path)
		require.NoError(t, err)
		defer db.Close()
		return pragma(t, db, "journal_mode")
	}

	t.Run("no WAL files are left behind", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.db")
		queries, db, err := database.Open(ctx, path)
		require.NoError(t, err)

		hash := fromHex(t, "cafe5"+strings.Repeat("0", 35))
		require.NoError(t, queries.InsertHash(ctx, sqlite.InsertHashParams{Hash: hash, Count: 7}))
		require.NoError(t, database.Close(ctx, db))
		assert.NoFileExists(t, path+"-wal")
		assert.NoFileExists(t, path+"-shm")

		// The committed row survived the checkpoint the mode switch performs
		queries, db, err = database.OpenReadOnly(ctx, path)
		require.NoError(t, err)
		count, err := queries.GetHashCount(ctx, hash)
		require.NoError(t, err)
		assert.Equal(t, int64(7), count)
		require.NoError(t, db.Close())

		// A read-only connection to a rollback-journal database creates neither
		assert.NoFileExists(t, path+"-wal")
		assert.NoFileExists(t, path+"-shm")
	})

	t.Run("switches mode despite a cancelled context", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "test.db")
		cancelled, cancel := context.WithCancel(ctx)
		_, db, err := database.Open(cancelled, path)
		require.NoError(t, err)

		cancel()
		require.NoError(t, database.Close(cancelled, db))
		assert.Equal(t, "delete", journalMode(t, path))
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
