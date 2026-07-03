package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"pwncache/database/sqlite"
)

// insertChunkSize is the number of rows bound to the reusable INSERT statement.
// Larger chunks amortise per-statement overhead but pay the modernc driver's
// per-parameter bind cost, which grows quadratically with statement size; see
// BenchmarkChunkSizes for the measured balance point.
const insertChunkSize = 100

// A HashInserter writes hash rows in bulk using multi-row INSERT statements.
//
// Full chunks of rows run through a statement prepared once at construction;
// any remainder runs through an exact-size statement built per call. sqlc
// cannot generate multi-row inserts for SQLite, hence this hand-written type.
type HashInserter struct {
	chunkSize      int
	chunkStatement *sql.Stmt
}

// NewHashInserter prepares the chunk statement against db.
// Close the returned HashInserter once finished with it.
func NewHashInserter(ctx context.Context, db *sql.DB) (*HashInserter, error) {
	return newHashInserter(ctx, db, insertChunkSize)
}

// newHashInserter lets benchmarks vary the chunk size.
func newHashInserter(ctx context.Context, db *sql.DB, chunkSize int) (*HashInserter, error) {
	statement, err := db.PrepareContext(ctx, insertHashesSQL(chunkSize))
	if err != nil {
		return nil, fmt.Errorf("preparing bulk insert: %w", err)
	}
	return &HashInserter{chunkSize: chunkSize, chunkStatement: statement}, nil
}

// Close releases the prepared chunk statement.
func (h *HashInserter) Close() error {
	return h.chunkStatement.Close()
}

// Insert writes every row inside the given transaction.
// Rows are not deduplicated: a hash that already exists in the table, or
// appears twice in rows, fails the whole call with a constraint error.
func (h *HashInserter) Insert(ctx context.Context, tx *sql.Tx, rows []sqlite.InsertHashParams) error {
	arguments := make([]any, 0, 2*min(len(rows), h.chunkSize))

	if len(rows) >= h.chunkSize {
		// One transaction-scoped statement serves every full chunk
		statement := tx.StmtContext(ctx, h.chunkStatement)
		for len(rows) >= h.chunkSize {
			arguments = bindHashes(arguments, rows[:h.chunkSize])
			rows = rows[h.chunkSize:]
			if _, err := statement.ExecContext(ctx, arguments...); err != nil {
				return fmt.Errorf("inserting hashes: %w", err)
			}
		}
	}

	if len(rows) == 0 {
		return nil
	}
	arguments = bindHashes(arguments, rows)
	if _, err := tx.ExecContext(ctx, insertHashesSQL(len(rows)), arguments...); err != nil {
		return fmt.Errorf("inserting hashes: %w", err)
	}
	return nil
}

// bindHashes flattens rows into bind arguments, reusing the given backing slice.
func bindHashes(arguments []any, rows []sqlite.InsertHashParams) []any {
	arguments = arguments[:0]
	for _, row := range rows {
		arguments = append(arguments, row.Hash, row.Count)
	}
	return arguments
}

// insertHashesSQL builds a multi-row INSERT statement for the given row count.
func insertHashesSQL(rows int) string {
	var builder strings.Builder
	builder.Grow(48 + 6*rows)
	builder.WriteString("INSERT INTO hashes (hash, count) VALUES ")
	for i := range rows {
		if i > 0 {
			builder.WriteByte(',')
		}
		builder.WriteString("(?,?)")
	}
	return builder.String()
}
