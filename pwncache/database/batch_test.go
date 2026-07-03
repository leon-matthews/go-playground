package database_test

import (
	"context"
	"database/sql"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pwncache/database"
	"pwncache/database/sqlite"
)

// syntheticRows builds distinct 20-byte hashes with predictable counts.
// Hashes are numbered from offset so successive calls never collide.
func syntheticRows(offset, count int) []sqlite.InsertHashParams {
	rows := make([]sqlite.InsertHashParams, count)
	for i := range count {
		hash := make([]byte, 20)
		binary.BigEndian.PutUint64(hash[12:], uint64(offset+i))
		rows[i] = sqlite.InsertHashParams{Hash: hash, Count: int64(offset + i + 1)}
	}
	return rows
}

// insertAll runs one Insert call inside a committed transaction.
func insertAll(
	t *testing.T,
	db *sql.DB,
	inserter *database.HashInserter,
	rows []sqlite.InsertHashParams,
) error {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)
	if err := inserter.Insert(ctx, tx, rows); err != nil {
		require.NoError(t, tx.Rollback())
		return err
	}
	return tx.Commit()
}

// countHashes returns how many rows the hashes table holds.
func countHashes(t *testing.T, db *sql.DB) int {
	t.Helper()
	var count int
	err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM hashes").Scan(&count)
	require.NoError(t, err)
	return count
}

func TestHashInserter(t *testing.T) {
	ctx := context.Background()

	// Sizes chosen around the 100-row chunk inside HashInserter
	cases := []struct {
		name string
		rows int
	}{
		{"empty", 0},
		{"single row", 1},
		{"partial chunk", 70},
		{"exact chunk", 100},
		{"chunks plus tail", 2_317},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			queries, db := open(t)
			inserter, err := database.NewHashInserter(ctx, db)
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, inserter.Close()) })

			rows := syntheticRows(0, tc.rows)
			require.NoError(t, insertAll(t, db, inserter, rows))
			assert.Equal(t, tc.rows, countHashes(t, db))

			// Every row landed with its own count, not a neighbour's
			for _, row := range rows {
				count, err := queries.GetHashCount(ctx, row.Hash)
				require.NoError(t, err)
				assert.Equal(t, row.Count, count)
			}
		})
	}

	t.Run("duplicate hash fails whole call", func(t *testing.T) {
		_, db := open(t)
		inserter, err := database.NewHashInserter(ctx, db)
		require.NoError(t, err)
		t.Cleanup(func() { assert.NoError(t, inserter.Close()) })

		rows := syntheticRows(0, 10)
		rows[9] = rows[0]
		require.ErrorContains(t, insertAll(t, db, inserter, rows), "UNIQUE constraint")
		assert.Equal(t, 0, countHashes(t, db))
	})
}

// BenchmarkInsertHashes compares one prefix's worth of single-row inserts
// against the batched equivalent, each iteration in its own transaction.
func BenchmarkInsertHashes(b *testing.B) {
	const rowsPerPrefix = 1_750
	ctx := context.Background()

	b.Run("single row", func(b *testing.B) {
		queries, db := open(b)
		next := 0
		for b.Loop() {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				b.Fatal(err)
			}
			qtx := queries.WithTx(tx)
			for _, row := range syntheticRows(next, rowsPerPrefix) {
				if err := qtx.InsertHash(ctx, row); err != nil {
					b.Fatal(err)
				}
			}
			if err := tx.Commit(); err != nil {
				b.Fatal(err)
			}
			next += rowsPerPrefix
		}
	})

	b.Run("batched", func(b *testing.B) {
		_, db := open(b)
		inserter, err := database.NewHashInserter(ctx, db)
		if err != nil {
			b.Fatal(err)
		}
		defer inserter.Close()

		next := 0
		for b.Loop() {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				b.Fatal(err)
			}
			if err := inserter.Insert(ctx, tx, syntheticRows(next, rowsPerPrefix)); err != nil {
				b.Fatal(err)
			}
			if err := tx.Commit(); err != nil {
				b.Fatal(err)
			}
			next += rowsPerPrefix
		}
	})
}
