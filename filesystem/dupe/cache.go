package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// openTimeout bounds the initial sql.DB.Ping so an unreachable file fails fast.
const openTimeout = 5 * time.Second

// flushSize commits the writer's pending transaction once this many rows queue up.
const flushSize = 10000

// flushInterval forces a commit if no row arrives within this window.
const flushInterval = 1 * time.Second

// writeChanBuffer sizes the worker→writer channel; full means backpressure on Set.
const writeChanBuffer = 64

// sqliteMagic is the 16-byte file header that identifies a SQLite 3 database.
const sqliteMagic = "SQLite format 3\x00"

// errClosed signals that the cache shut down before an op was accepted.
var errClosed = errors.New("cache closed")

// CacheEntry is the persisted hash record for a single file; ModTime+Size are
// the verification fields used to invalidate stale entries.
type CacheEntry struct {
	Size    int64
	ModTime time.Time
	Hash    string
}

// Cache is a SQLite-backed hash cache with an async write pipeline. A Cache
// with a nil db is a no-op: Get always misses, Set/Sweep/Flush silently succeed.
type Cache struct {
	db        *sql.DB
	log       *slog.Logger
	writes    chan cacheOp
	done      chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
}

// cacheOp multiplexes the writer goroutine's input; exactly one field is set.
type cacheOp struct {
	write *writePayload
	sweep *sweepPayload
	flush *flushPayload
}

type writePayload struct {
	path  string
	entry CacheEntry
}

type sweepPayload struct {
	seen  map[string]struct{}
	roots []string
	done  chan error
}

type flushPayload struct {
	done chan struct{}
}

// cachePath returns the absolute path of the persistent hash cache.
func cachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "dupe", "cache.db"), nil
}

// openCache opens or creates the SQLite cache at path, replacing the file if
// it exists but is not a SQLite database.
func openCache(path string, log *slog.Logger) (*Cache, error) {
	return openCacheWithTimeout(path, openTimeout, log)
}

// openCacheWithTimeout is openCache with an explicit ping timeout for tests.
func openCacheWithTimeout(path string, timeout time.Duration, log *slog.Logger) (*Cache, error) {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	if path == "" {
		return &Cache{log: log}, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return &Cache{log: log}, err
	}
	if err := ensureSQLiteFile(path, log); err != nil {
		return &Cache{log: log}, err
	}

	db, err := sql.Open("sqlite", sqliteDSN(path))
	if err != nil {
		return &Cache{log: log}, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return &Cache{log: log}, err
	}

	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS hashes (
		path     TEXT    PRIMARY KEY,
		size     INTEGER NOT NULL,
		modtime  INTEGER NOT NULL,
		hash     BLOB    NOT NULL
	) WITHOUT ROWID`); err != nil {
		db.Close()
		return &Cache{log: log}, err
	}

	c := &Cache{
		db:     db,
		log:    log,
		writes: make(chan cacheOp, writeChanBuffer),
		done:   make(chan struct{}),
	}
	c.wg.Add(1)
	go c.writer()
	log.Debug("cache opened", "path", path)
	return c, nil
}

// sqliteDSN builds a modernc.org/sqlite DSN with our pragma bundle.
func sqliteDSN(path string) string {
	return fmt.Sprintf(
		"file:%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=temp_store(MEMORY)&_pragma=cache_size(-65536)",
		path,
	)
}

// ensureSQLiteFile checks an existing file's magic header and deletes it if
// it isn't a SQLite database, so the subsequent open recreates from scratch.
func ensureSQLiteFile(path string, log *slog.Logger) error {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	header := make([]byte, len(sqliteMagic))
	n, _ := f.Read(header)
	f.Close()
	if n == len(sqliteMagic) && string(header) == sqliteMagic {
		return nil
	}
	log.Warn("cache file is not a SQLite database; recreating", "path", path)
	return os.Remove(path)
}

// Close stops the writer goroutine (draining pending ops first) and closes the
// underlying database. Safe to call on a no-op Cache or to call repeatedly.
func (c *Cache) Close() error {
	if c == nil || c.db == nil {
		return nil
	}
	var err error
	c.closeOnce.Do(func() {
		close(c.done)
		c.wg.Wait()
		err = c.db.Close()
	})
	return err
}

// Get returns the cache entry for path, if present.
func (c *Cache) Get(path string) (CacheEntry, bool) {
	if c == nil || c.db == nil {
		return CacheEntry{}, false
	}
	var (
		size    int64
		modtime int64
		digest  []byte
	)
	err := c.db.QueryRow(
		"SELECT size, modtime, hash FROM hashes WHERE path = ?",
		path,
	).Scan(&size, &modtime, &digest)
	if errors.Is(err, sql.ErrNoRows) {
		return CacheEntry{}, false
	}
	if err != nil {
		c.log.Warn("cache: select failed; treating as miss", "path", path, "err", err)
		return CacheEntry{}, false
	}
	return CacheEntry{
		Size:    size,
		ModTime: time.Unix(0, modtime).UTC(),
		Hash:    hex.EncodeToString(digest),
	}, true
}

// Set queues an entry for asynchronous persistence. Returns nil on the happy
// path; per-row errors are logged by the writer.
func (c *Cache) Set(path string, entry CacheEntry) error {
	if c == nil || c.db == nil {
		return nil
	}
	op := cacheOp{write: &writePayload{path: path, entry: entry}}
	select {
	case c.writes <- op:
		return nil
	case <-c.done:
		return errClosed
	}
}

// Flush blocks until all writes queued before this call have been committed.
func (c *Cache) Flush() error {
	if c == nil || c.db == nil {
		return nil
	}
	ack := make(chan struct{}, 1)
	select {
	case c.writes <- cacheOp{flush: &flushPayload{done: ack}}:
	case <-c.done:
		return errClosed
	}
	select {
	case <-ack:
		return nil
	case <-c.done:
		return errClosed
	}
}

// Sweep removes entries under any of roots that aren't in seen. Blocks until
// the writer applies the operation.
func (c *Cache) Sweep(seen map[string]struct{}, roots []string) error {
	if c == nil || c.db == nil {
		return nil
	}
	ack := make(chan error, 1)
	op := cacheOp{sweep: &sweepPayload{seen: seen, roots: roots, done: ack}}
	select {
	case c.writes <- op:
	case <-c.done:
		return errClosed
	}
	select {
	case err := <-ack:
		return err
	case <-c.done:
		return errClosed
	}
}

// writer owns the write connection and serialises all mutating ops.
func (c *Cache) writer() {
	defer c.wg.Done()

	var (
		tx      *sql.Tx
		stmt    *sql.Stmt
		pending int
	)

	openTx := func() error {
		t, err := c.db.Begin()
		if err != nil {
			return err
		}
		s, err := t.Prepare("INSERT OR REPLACE INTO hashes(path, size, modtime, hash) VALUES (?, ?, ?, ?)")
		if err != nil {
			t.Rollback()
			return err
		}
		tx, stmt, pending = t, s, 0
		return nil
	}

	commitTx := func() {
		if tx == nil {
			return
		}
		stmt.Close()
		if err := tx.Commit(); err != nil {
			c.log.Warn("cache: commit failed", "err", err)
			tx.Rollback()
		}
		tx, stmt, pending = nil, nil, 0
	}

	handleWrite := func(p *writePayload) {
		digest, err := hex.DecodeString(p.entry.Hash)
		if err != nil || len(digest) != sha256.Size {
			c.log.Warn("cache: invalid hash; skipping", "path", p.path, "err", err)
			return
		}
		if tx == nil {
			if err := openTx(); err != nil {
				c.log.Warn("cache: begin failed", "err", err)
				return
			}
		}
		if _, err := stmt.Exec(p.path, p.entry.Size, p.entry.ModTime.UnixNano(), digest); err != nil {
			c.log.Warn("cache: insert failed", "path", p.path, "err", err)
			return
		}
		pending++
		if pending >= flushSize {
			commitTx()
		}
	}

	handle := func(op cacheOp) {
		switch {
		case op.write != nil:
			handleWrite(op.write)
		case op.flush != nil:
			commitTx()
			op.flush.done <- struct{}{}
		case op.sweep != nil:
			commitTx()
			op.sweep.done <- c.runSweep(op.sweep.seen, op.sweep.roots)
		}
	}

	for {
		timer := time.NewTimer(flushInterval)
		select {
		case op := <-c.writes:
			timer.Stop()
			handle(op)
		case <-timer.C:
			commitTx()
		case <-c.done:
			timer.Stop()
			for {
				select {
				case op := <-c.writes:
					handle(op)
				default:
					commitTx()
					return
				}
			}
		}
	}
}

// runSweep executes the sweep on a fresh transaction; called by the writer.
func (c *Cache) runSweep(seen map[string]struct{}, roots []string) error {
	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("begin sweep tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("CREATE TEMP TABLE seen (path TEXT PRIMARY KEY)"); err != nil {
		return fmt.Errorf("create temp table: %w", err)
	}

	ins, err := tx.Prepare("INSERT INTO seen VALUES (?)")
	if err != nil {
		return fmt.Errorf("prepare seen insert: %w", err)
	}
	for path := range seen {
		if _, err := ins.Exec(path); err != nil {
			ins.Close()
			return fmt.Errorf("insert seen: %w", err)
		}
	}
	ins.Close()

	del, err := tx.Prepare(`DELETE FROM hashes
		WHERE (path = ?1 OR substr(path, 1, length(?1)+1) = ?1 || '/')
		  AND path NOT IN (SELECT path FROM seen)`)
	if err != nil {
		return fmt.Errorf("prepare delete: %w", err)
	}
	for _, root := range roots {
		if _, err := del.Exec(root); err != nil {
			del.Close()
			return fmt.Errorf("delete sweep: %w", err)
		}
	}
	del.Close()

	if _, err := tx.Exec("DROP TABLE seen"); err != nil {
		return fmt.Errorf("drop temp table: %w", err)
	}
	return tx.Commit()
}

// pathInRoots reports whether path is equal to, or nested under, any of roots.
func pathInRoots(path string, roots []string) bool {
	sep := string(filepath.Separator)
	for _, r := range roots {
		if path == r || strings.HasPrefix(path, r+sep) {
			return true
		}
	}
	return false
}
