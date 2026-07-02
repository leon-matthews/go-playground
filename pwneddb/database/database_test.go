package database_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pwneddb/database"
	"pwneddb/database/sqlite"
)

func TestOpenAndUpsert(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "test.db")

	queries, db, err := database.Open(ctx, path)
	require.NoError(t, err)
	defer db.Close()

	insert := sqlite.UpsertPrefixParams{
		Prefix:  "cafe5",
		Updated: sql.NullInt64{Int64: 100, Valid: true},
		Etag:    sql.NullString{String: "etag-1", Valid: true},
		Hashes:  "hash-list-1",
	}
	require.NoError(t, queries.UpsertPrefix(ctx, insert))

	row, err := queries.GetPrefix(ctx, "cafe5")
	require.NoError(t, err)
	assert.Equal(t, "hash-list-1", row.Hashes)
	assert.Equal(t, "etag-1", row.Etag.String)

	// Upserting the same prefix updates the row in place
	update := sqlite.UpsertPrefixParams{
		Prefix:  "cafe5",
		Updated: sql.NullInt64{Int64: 200, Valid: true},
		Etag:    sql.NullString{String: "etag-2", Valid: true},
		Hashes:  "hash-list-2",
	}
	require.NoError(t, queries.UpsertPrefix(ctx, update))

	row, err = queries.GetPrefix(ctx, "cafe5")
	require.NoError(t, err)
	assert.Equal(t, "hash-list-2", row.Hashes)
	assert.Equal(t, "etag-2", row.Etag.String)

	// The conflict updated rather than inserted, so still one row
	etags, err := queries.GetEtags(ctx)
	require.NoError(t, err)
	assert.Len(t, etags, 1)
}
