package database

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBatchWriter(t *testing.T) {
	ctx := context.Background()

	t.Run("reports rows changed like a direct upsert", func(t *testing.T) {
		_, db, err := Open(ctx, filepath.Join(t.TempDir(), "batch.db"))
		require.NoError(t, err)
		defer db.Close()
		bw := NewBatchWriter(db)

		changed, err := bw.Upsert(ctx, "hunter2", 42)
		require.NoError(t, err)
		assert.Equal(t, int64(1), changed, "new row")

		changed, err = bw.Upsert(ctx, "hunter2", 42)
		require.NoError(t, err)
		assert.Equal(t, int64(0), changed, "same count changes nothing")

		changed, err = bw.Upsert(ctx, "hunter2", 99)
		require.NoError(t, err)
		assert.Equal(t, int64(1), changed, "new count is a change")

		// A fresh query only sees the rows once close commits the open batch.
		require.NoError(t, bw.Close())
		var count, rows int64
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count FROM passwords WHERE password=?", "hunter2").Scan(&count))
		assert.Equal(t, int64(99), count)
		require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM passwords").Scan(&rows))
		assert.Equal(t, int64(1), rows)
	})

	t.Run("cancellation is not fatal and close still commits", func(t *testing.T) {
		_, db, err := Open(ctx, filepath.Join(t.TempDir(), "cancel.db"))
		require.NoError(t, err)
		defer db.Close()
		bw := NewBatchWriter(db)

		_, err = bw.Upsert(ctx, "keepme", 5)
		require.NoError(t, err)

		cctx, cancel := context.WithCancel(ctx)
		cancel()
		_, err = bw.Upsert(cctx, "dropme", 9)
		assert.ErrorIs(t, err, context.Canceled)

		require.NoError(t, bw.Close())
		var count int64
		require.NoError(t, db.QueryRowContext(ctx, "SELECT count FROM passwords WHERE password=?", "keepme").Scan(&count))
		assert.Equal(t, int64(5), count, "pre-cancellation row survives")

		err = db.QueryRowContext(ctx, "SELECT count FROM passwords WHERE password=?", "dropme").Scan(&count)
		assert.ErrorIs(t, err, sql.ErrNoRows, "cancelled upsert did not persist")
	})

	t.Run("is safe under concurrent writers", func(t *testing.T) {
		_, db, err := Open(ctx, filepath.Join(t.TempDir(), "concurrent.db"))
		require.NoError(t, err)
		defer db.Close()
		bw := NewBatchWriter(db)

		// workers*each exceeds batchMaxRows, so a count-triggered commit fires
		// mid-run while other goroutines are upserting - the path -race must clear.
		const workers, each = 8, 2000
		var wg sync.WaitGroup
		for w := range workers {
			wg.Go(func() {
				for i := range each {
					if _, err := bw.Upsert(ctx, fmt.Sprintf("w%d-%d", w, i), int64(i)); err != nil {
						t.Errorf("upsert: %v", err)
						return
					}
				}
			})
		}
		wg.Wait()
		require.NoError(t, bw.Close())

		var rows int64
		require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM passwords").Scan(&rows))
		assert.Equal(t, int64(workers*each), rows)
	})
}
