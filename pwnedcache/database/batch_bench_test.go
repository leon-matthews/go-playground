package database

import (
	"context"
	"encoding/binary"
	"fmt"
	"path/filepath"
	"testing"

	"pwnedcache/database/sqlite"
)

// BenchmarkChunkSizes measures Insert throughput across chunk sizes, looking
// for the balance between per-statement overhead, which favours large chunks,
// and the driver's quadratic per-parameter bind cost, which favours small ones.
func BenchmarkChunkSizes(b *testing.B) {
	const rowsPerPrefix = 1_750
	ctx := context.Background()

	for _, chunkSize := range []int{1, 10, 25, 50, 100, 250, 500, 1000} {
		b.Run(fmt.Sprintf("chunk-%d", chunkSize), func(b *testing.B) {
			_, db, err := Open(ctx, filepath.Join(b.TempDir(), "bench.db"))
			if err != nil {
				b.Fatal(err)
			}
			defer db.Close()
			inserter, err := newHashInserter(ctx, db, chunkSize)
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
				if err := inserter.Insert(ctx, tx, benchmarkRows(next, rowsPerPrefix)); err != nil {
					b.Fatal(err)
				}
				if err := tx.Commit(); err != nil {
					b.Fatal(err)
				}
				next += rowsPerPrefix
			}
		})
	}
}

// benchmarkRows builds distinct 20-byte hashes, numbered from offset.
func benchmarkRows(offset, count int) []sqlite.InsertHashParams {
	rows := make([]sqlite.InsertHashParams, count)
	for i := range count {
		hash := make([]byte, 20)
		binary.BigEndian.PutUint64(hash[12:], uint64(offset+i))
		rows[i] = sqlite.InsertHashParams{Hash: hash, Count: int64(offset + i + 1)}
	}
	return rows
}
