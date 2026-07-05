package database

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pwnedpasswords/database/sqlite"
)

func TestUpsertPassword(t *testing.T) {
	ctx := context.Background()
	queries, db, err := Open(ctx, filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	defer db.Close()

	upsert := func(password string, count int64) int64 {
		rows, err := queries.UpsertPassword(ctx, sqlite.UpsertPasswordParams{Password: password, Count: count})
		require.NoError(t, err)
		return rows
	}

	// UpsertPassword reports one changed row when it adds or updates, and zero
	// when the password already exists with the same count.
	t.Run("a new password is one change", func(t *testing.T) {
		assert.Equal(t, int64(1), upsert("hunter2", 42))
	})

	t.Run("re-inserting the same count changes nothing", func(t *testing.T) {
		assert.Equal(t, int64(0), upsert("hunter2", 42))
	})

	t.Run("a new count is one change", func(t *testing.T) {
		assert.Equal(t, int64(1), upsert("hunter2", 99))
	})

	t.Run("a second password is one change", func(t *testing.T) {
		assert.Equal(t, int64(1), upsert("letmein", 7))
	})
}
