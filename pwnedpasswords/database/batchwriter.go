package database

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"pwnedpasswords/database/sqlite"
)

// Batch commit thresholds: an open batch is committed once it reaches batchMaxRows
// upserts or batchMaxAge of wall-clock time, whichever comes first. batchTick is
// how often the background flusher checks the age, so a sparse stream (bruteforce)
// still commits promptly instead of holding a transaction open indefinitely.
const (
	batchMaxRows = 10_000
	batchMaxAge  = 60 * time.Second
	batchTick    = 5 * time.Second
)

// BatchWriter groups password upserts into transactions, committing when a batch
// fills or ages out and checkpointing the WAL after each commit so it stays
// bounded. Its methods are safe for concurrent use by the bruteforce workers; the
// mutex serialises the single write connection that autocommit relied on before.
type BatchWriter struct {
	db *sql.DB

	mu   sync.Mutex
	tx   *sql.Tx
	q    *sqlite.Queries
	rows int
	open time.Time
	err  error // first non-cancellation error; sticky once set

	stop chan struct{}
	done chan struct{}
}

// NewBatchWriter returns a writer with a running flusher that ages out batches.
// The caller must call Close to commit the final batch and stop the flusher.
func NewBatchWriter(db *sql.DB) *BatchWriter {
	b := &BatchWriter{
		db:   db,
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
	go b.flushLoop()
	return b
}

// Upsert records one password, opening a batch if none is in progress and
// committing once the batch is full. It returns the rows changed, exactly as
// [sqlite.Queries.UpsertPassword] reports, so callers keep their found/changed
// accounting unchanged.
func (b *BatchWriter) Upsert(ctx context.Context, password string, count int64) (int64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.err != nil {
		return 0, b.err
	}

	if b.tx == nil {
		tx, err := b.db.BeginTx(ctx, nil)
		if err != nil {
			return 0, b.recordErr(err)
		}
		b.tx, b.q, b.rows, b.open = tx, sqlite.New(tx), 0, time.Now()
	}

	changed, err := b.q.UpsertPassword(ctx, sqlite.UpsertPasswordParams{Password: password, Count: count})
	if err != nil {
		return 0, b.recordErr(err)
	}
	b.rows++
	if b.rows >= batchMaxRows {
		if err := b.commitLocked(); err != nil {
			return 0, b.recordErr(err)
		}
	}
	return changed, nil
}

// Close stops the flusher, commits any pending batch, and truncates the WAL.
func (b *BatchWriter) Close() error {
	close(b.stop)
	<-b.done

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.err != nil {
		if b.tx != nil {
			_ = b.tx.Rollback()
		}
		return b.err
	}
	if err := b.commitLocked(); err != nil {
		return b.recordErr(err)
	}
	_, err := b.db.ExecContext(context.Background(), "PRAGMA wal_checkpoint(TRUNCATE)")
	return err
}

// flushLoop commits the open batch once it has aged past batchMaxAge.
func (b *BatchWriter) flushLoop() {
	defer close(b.done)
	ticker := time.NewTicker(batchTick)
	defer ticker.Stop()
	for {
		select {
		case <-b.stop:
			return
		case <-ticker.C:
			b.mu.Lock()
			if b.err == nil && b.tx != nil && time.Since(b.open) >= batchMaxAge {
				if err := b.commitLocked(); err != nil {
					b.recordErr(err)
				}
			}
			b.mu.Unlock()
		}
	}
}

// commitLocked commits the current batch and checkpoints the WAL, leaving no
// batch open. It is a no-op when none is in progress. The caller must hold b.mu.
func (b *BatchWriter) commitLocked() error {
	if b.tx == nil {
		return nil
	}
	tx := b.tx
	b.tx, b.q, b.rows = nil, nil, 0
	if err := tx.Commit(); err != nil {
		return err
	}
	// A passive checkpoint after each commit keeps the WAL from growing without
	// bound the way the continuous-autocommit writer let it (a 22 GiB WAL).
	_, err := b.db.ExecContext(context.Background(), "PRAGMA wal_checkpoint(PASSIVE)")
	return err
}

// recordErr abandons the open batch and, for a genuine (non-cancellation) error,
// stores it as sticky so later calls and close surface it. Cancellation is the
// expected interrupt path, so the batch is left for close to commit. It returns
// err unchanged. The caller must hold b.mu.
func (b *BatchWriter) recordErr(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}
	if b.tx != nil {
		_ = b.tx.Rollback()
		b.tx, b.q, b.rows = nil, nil, 0
	}
	if b.err == nil {
		b.err = err
	}
	return err
}
