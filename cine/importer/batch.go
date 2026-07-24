// Package importer builds a cine SQLite database from the IMDb dataset files.
package importer

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// bindParamBudget caps the bind parameters in one multi-row INSERT. The modernc
// driver's per-parameter bind cost rises sharply with statement size, so small
// chunks win: BenchmarkInsertChunkSizes peaks near 33 parameters (about 3 rows of
// the 11-column titles) and falls off steeply past it. Chunks are therefore sized
// to this budget (rows = budget / columns), not to a large fixed row count.
const bindParamBudget = 36

// batchInserter buffers rows for one table and writes them with multi-row INSERT
// statements.
//
// Full chunks run through a statement prepared once; the trailing remainder runs
// through an exact-size statement built for it. sqlc cannot generate multi-row
// inserts for SQLite, so this is hand-written.
type batchInserter[T any] struct {
	table     string
	columns   []string
	bind      func([]any, T) []any
	chunkRows int
	tx        *sql.Tx
	chunkStmt *sql.Stmt

	buffer []T
	args   []any
	added  int64
}

// newBatchInserter prepares the reusable chunk statement on tx, sizing the chunk
// from the shared bind-parameter budget.
func newBatchInserter[T any](ctx context.Context, tx *sql.Tx, table string, columns []string, bind func([]any, T) []any) (*batchInserter[T], error) {
	return newSizedInserter(ctx, tx, table, columns, bind, max(1, bindParamBudget/len(columns)))
}

// newSizedInserter is newBatchInserter with an explicit chunk size, for benchmarks.
func newSizedInserter[T any](ctx context.Context, tx *sql.Tx, table string, columns []string, bind func([]any, T) []any, chunkRows int) (*batchInserter[T], error) {
	stmt, err := tx.PrepareContext(ctx, insertSQL(table, columns, chunkRows))
	if err != nil {
		return nil, fmt.Errorf("preparing %s insert: %w", table, err)
	}
	return &batchInserter[T]{
		table:     table,
		columns:   columns,
		bind:      bind,
		chunkRows: chunkRows,
		tx:        tx,
		chunkStmt: stmt,
		buffer:    make([]T, 0, chunkRows),
		args:      make([]any, 0, chunkRows*len(columns)),
	}, nil
}

// Add buffers one row, flushing a full chunk once enough have accumulated.
func (b *batchInserter[T]) Add(ctx context.Context, row T) error {
	b.buffer = append(b.buffer, row)
	b.added++
	if len(b.buffer) < b.chunkRows {
		return nil
	}
	b.args = b.bindRows(b.buffer)
	if _, err := b.chunkStmt.ExecContext(ctx, b.args...); err != nil {
		return fmt.Errorf("inserting %s: %w", b.table, err)
	}
	b.buffer = b.buffer[:0]
	return nil
}

// Flush writes any buffered rows left over from the last partial chunk.
func (b *batchInserter[T]) Flush(ctx context.Context) error {
	if len(b.buffer) == 0 {
		return nil
	}
	b.args = b.bindRows(b.buffer)
	query := insertSQL(b.table, b.columns, len(b.buffer))
	if _, err := b.tx.ExecContext(ctx, query, b.args...); err != nil {
		return fmt.Errorf("inserting %s: %w", b.table, err)
	}
	b.buffer = b.buffer[:0]
	return nil
}

// Added reports how many rows have been passed to Add over this inserter's life.
func (b *batchInserter[T]) Added() int64 {
	return b.added
}

// bindRows flattens the buffered rows into one bind-argument slice, reusing the
// backing array.
func (b *batchInserter[T]) bindRows(rows []T) []any {
	args := b.args[:0]
	for _, row := range rows {
		args = b.bind(args, row)
	}
	return args
}

// insertSQL builds a multi-row INSERT for the given number of value groups.
func insertSQL(table string, columns []string, rows int) string {
	group := "(" + strings.TrimSuffix(strings.Repeat("?,", len(columns)), ",") + ")"

	var b strings.Builder
	b.Grow(len(table) + len(columns)*12 + rows*(len(group)+1))
	b.WriteString("INSERT INTO ")
	b.WriteString(table)
	b.WriteString(" (")
	b.WriteString(strings.Join(columns, ", "))
	b.WriteString(") VALUES ")
	for i := range rows {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(group)
	}
	return b.String()
}
