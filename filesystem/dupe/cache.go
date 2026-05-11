package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
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

// schemaVersion is bumped whenever the on-disk schema changes; mismatch triggers a rebuild.
const schemaVersion = 2

// errClosed signals that the cache shut down before an op was accepted.
var errClosed = errors.New("cache closed")

// CacheEntry is the persisted hash record for a single file; ModTime+Size are the verification fields.
type CacheEntry struct {
	Size    int64
	ModTime time.Time
	Hash    [32]byte
}

// Cache is a SQLite-backed hash cache with an async write pipeline. A Cache with a nil db is a no-op.
type Cache struct {
	db                   *sql.DB
	getStmt              *sql.Stmt
	getFolderMtimeStmt   *sql.Stmt
	getFilesInFolderStmt *sql.Stmt
	log                  *slog.Logger
	writes               chan cacheOp
	done                 chan struct{}
	closeOnce            sync.Once
	wg                   sync.WaitGroup
}

// cacheOp multiplexes the writer goroutine's input; exactly one field is set.
type cacheOp struct {
	write          *writePayload
	flush          *flushPayload
	sweep          *sweepPayload
	setFolderMtime *folderMtimePayload
}

type writePayload struct {
	path  string
	entry CacheEntry
}

type sweepPayload struct {
	seenFolders map[string]struct{}
	seenFiles   map[string]map[string]struct{}
	roots       []string
	done        chan error
}

type flushPayload struct {
	done chan struct{}
}

type folderMtimePayload struct {
	path  string
	mtime time.Time
}

// cachePath returns the absolute path of the persistent hash cache.
func cachePath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "dupe", "cache.db"), nil
}

// openCache opens or creates the SQLite cache at path; rebuilds the schema if it's stale or absent.
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
	log.Debug("openCache: starting", "path", path, "timeout", timeout)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return &Cache{log: log}, err
	}
	if err := ensureSQLiteFile(path, log); err != nil {
		return &Cache{log: log}, err
	}

	_, statErr := os.Stat(path)
	freshFile := errors.Is(statErr, os.ErrNotExist)
	log.Debug("openCache: file existence checked", "fresh", freshFile)

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
	log.Debug("openCache: db connected")

	if err := ensureSchema(db, log); err != nil {
		db.Close()
		return &Cache{log: log}, err
	}
	log.Debug("openCache: schema ensured")
	if freshFile {
		log.Info("new cache file created", "path", path)
	}

	getStmt, err := db.Prepare(`SELECT f.size, f.modtime, f.hash
		FROM files f JOIN folders d ON d.id = f.folder_id
		WHERE d.path = ? AND f.name = ?`)
	if err != nil {
		db.Close()
		return &Cache{log: log}, err
	}
	getFolderMtimeStmt, err := db.Prepare("SELECT mtime FROM folders WHERE path = ?")
	if err != nil {
		getStmt.Close()
		db.Close()
		return &Cache{log: log}, err
	}
	getFilesInFolderStmt, err := db.Prepare(`SELECT f.name, f.size, f.modtime, f.hash
		FROM files f JOIN folders d ON d.id = f.folder_id
		WHERE d.path = ?`)
	if err != nil {
		getFolderMtimeStmt.Close()
		getStmt.Close()
		db.Close()
		return &Cache{log: log}, err
	}
	log.Debug("openCache: statements prepared")

	c := &Cache{
		db:                   db,
		getStmt:              getStmt,
		getFolderMtimeStmt:   getFolderMtimeStmt,
		getFilesInFolderStmt: getFilesInFolderStmt,
		log:                  log,
		writes:               make(chan cacheOp, writeChanBuffer),
		done:                 make(chan struct{}),
	}
	c.wg.Add(1)
	go c.writer()
	log.Debug("openCache: ready", "path", path)
	log.Debug("cache opened", "path", path)
	return c, nil
}

// sqliteDSN builds a modernc.org/sqlite DSN with our pragma bundle.
func sqliteDSN(path string) string {
	return fmt.Sprintf(
		"file:%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=temp_store(MEMORY)&_pragma=cache_size(-65536)&_pragma=foreign_keys(on)",
		path,
	)
}

// ensureSchema checks PRAGMA user_version and rebuilds the schema if it doesn't match.
func ensureSchema(db *sql.DB, log *slog.Logger) error {
	var version int
	if err := db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}
	if version == schemaVersion {
		return nil
	}
	if version != 0 {
		log.Warn("cache schema version mismatch; rebuilding", "had", version, "want", schemaVersion)
	}
	for _, table := range []string{"hashes", "files", "folders"} {
		if _, err := db.Exec("DROP TABLE IF EXISTS " + table); err != nil {
			return fmt.Errorf("drop %s: %w", table, err)
		}
	}
	if _, err := db.Exec(`CREATE TABLE folders (
		id    INTEGER PRIMARY KEY,
		path  TEXT    UNIQUE NOT NULL,
		mtime INTEGER NOT NULL
	)`); err != nil {
		return fmt.Errorf("create folders: %w", err)
	}
	if _, err := db.Exec(`CREATE TABLE files (
		folder_id INTEGER NOT NULL,
		name      TEXT    NOT NULL,
		size      INTEGER NOT NULL,
		modtime   INTEGER NOT NULL,
		hash      BLOB    NOT NULL,
		PRIMARY KEY (folder_id, name),
		FOREIGN KEY (folder_id) REFERENCES folders(id) ON DELETE CASCADE
	) WITHOUT ROWID`); err != nil {
		return fmt.Errorf("create files: %w", err)
	}
	if _, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", schemaVersion)); err != nil {
		return fmt.Errorf("set schema version: %w", err)
	}
	return nil
}

// ensureSQLiteFile deletes path if it exists but isn't a SQLite database.
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

// Close stops the writer goroutine and closes the underlying database. Idempotent.
func (c *Cache) Close() error {
	if c == nil || c.db == nil {
		return nil
	}
	var err error
	c.closeOnce.Do(func() {
		close(c.done)
		c.wg.Wait()
		c.getFilesInFolderStmt.Close()
		c.getFolderMtimeStmt.Close()
		c.getStmt.Close()
		err = c.db.Close()
	})
	return err
}

// Get returns the cache entry for path, if present.
func (c *Cache) Get(path string) (CacheEntry, bool) {
	if c == nil || c.db == nil {
		return CacheEntry{}, false
	}
	folder, name := splitFolderPath(path)
	var (
		size    int64
		modtime int64
		digest  []byte
	)
	err := c.getStmt.QueryRow(folder, name).Scan(&size, &modtime, &digest)
	if errors.Is(err, sql.ErrNoRows) {
		return CacheEntry{}, false
	}
	if err != nil {
		c.log.Warn("cache: select failed; treating as miss", "path", path, "err", err)
		return CacheEntry{}, false
	}
	if len(digest) != sha256.Size {
		c.log.Warn("cache: unexpected hash length; treating as miss", "path", path, "len", len(digest))
		return CacheEntry{}, false
	}
	entry := CacheEntry{
		Size:    size,
		ModTime: time.Unix(0, modtime).UTC(),
	}
	copy(entry.Hash[:], digest)
	return entry, true
}

// Set queues an entry for asynchronous persistence.
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
	c.log.Debug("flush: requested")
	ack := make(chan struct{}, 1)
	select {
	case c.writes <- cacheOp{flush: &flushPayload{done: ack}}:
	case <-c.done:
		return errClosed
	}
	select {
	case <-ack:
		c.log.Debug("flush: acknowledged")
		return nil
	case <-c.done:
		return errClosed
	}
}

// Sweep removes folders under any root that weren't visited (CASCADE drops their files), and
// drops files under visited folders whose basename isn't in the per-folder seen set. Builds
// the seen sets internally from the scanner's FolderScans.
func (c *Cache) Sweep(folderScans []FolderScan, roots []string) error {
	if c == nil || c.db == nil {
		return nil
	}
	seenFolders, seenFiles := buildSweepSets(folderScans)
	ack := make(chan error, 1)
	op := cacheOp{sweep: &sweepPayload{
		seenFolders: seenFolders,
		seenFiles:   seenFiles,
		roots:       roots,
		done:        ack,
	}}
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

// buildSweepSets folds FolderScans into the seenFolders/seenFiles map shape Sweep's internal
// SQL expects.
func buildSweepSets(folderScans []FolderScan) (map[string]struct{}, map[string]map[string]struct{}) {
	seenFolders := make(map[string]struct{}, len(folderScans))
	seenFiles := make(map[string]map[string]struct{}, len(folderScans))
	for _, fs := range folderScans {
		seenFolders[fs.Path] = struct{}{}
		names := make(map[string]struct{}, len(fs.Children))
		for _, name := range fs.Children {
			names[name] = struct{}{}
		}
		seenFiles[fs.Path] = names
	}
	return seenFolders, seenFiles
}

// GetFolderMtime returns the cached mtime for folder, if present.
func (c *Cache) GetFolderMtime(folder string) (time.Time, bool) {
	if c == nil || c.db == nil {
		return time.Time{}, false
	}
	var nanos int64
	err := c.getFolderMtimeStmt.QueryRow(folder).Scan(&nanos)
	if errors.Is(err, sql.ErrNoRows) {
		return time.Time{}, false
	}
	if err != nil {
		c.log.Warn("cache: folder mtime select failed", "folder", folder, "err", err)
		return time.Time{}, false
	}
	return time.Unix(0, nanos).UTC(), true
}

// SetFolderMtime queues an upsert of folder's mtime; async.
func (c *Cache) SetFolderMtime(folder string, mtime time.Time) error {
	if c == nil || c.db == nil {
		return nil
	}
	op := cacheOp{setFolderMtime: &folderMtimePayload{path: folder, mtime: mtime}}
	select {
	case c.writes <- op:
		return nil
	case <-c.done:
		return errClosed
	}
}

// GetFilesInFolder returns all cached files directly under folder, keyed by basename.
func (c *Cache) GetFilesInFolder(folder string) (map[string]CacheEntry, error) {
	if c == nil || c.db == nil {
		return nil, nil
	}
	rows, err := c.getFilesInFolderStmt.Query(folder)
	if err != nil {
		return nil, fmt.Errorf("query files in folder: %w", err)
	}
	defer rows.Close()
	out := make(map[string]CacheEntry)
	for rows.Next() {
		var (
			name    string
			size    int64
			modtime int64
			digest  []byte
		)
		if err := rows.Scan(&name, &size, &modtime, &digest); err != nil {
			return nil, fmt.Errorf("scan file row: %w", err)
		}
		if len(digest) != sha256.Size {
			c.log.Warn("cache: unexpected hash length; skipping", "folder", folder, "name", name, "len", len(digest))
			continue
		}
		entry := CacheEntry{Size: size, ModTime: time.Unix(0, modtime).UTC()}
		copy(entry.Hash[:], digest)
		out[name] = entry
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate file rows: %w", err)
	}
	return out, nil
}

// writer owns the write connection and serialises all mutating ops.
func (c *Cache) writer() {
	defer c.wg.Done()
	c.log.Debug("writer: started")

	var (
		tx                *sql.Tx
		folderEnsureStmt  *sql.Stmt
		folderMtimeStmt   *sql.Stmt
		fileStmt          *sql.Stmt
		pending           int
		idCache           = make(map[string]int64)
		commitCount       int
	)

	openTx := func() error {
		t, err := c.db.Begin()
		if err != nil {
			return err
		}
		fe, err := t.Prepare(`INSERT INTO folders(path, mtime) VALUES (?, 0)
			ON CONFLICT(path) DO UPDATE SET path = excluded.path
			RETURNING id`)
		if err != nil {
			t.Rollback()
			return err
		}
		fm, err := t.Prepare(`INSERT INTO folders(path, mtime) VALUES (?, ?)
			ON CONFLICT(path) DO UPDATE SET mtime = excluded.mtime
			RETURNING id`)
		if err != nil {
			fe.Close()
			t.Rollback()
			return err
		}
		fls, err := t.Prepare(`INSERT INTO files(folder_id, name, size, modtime, hash) VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(folder_id, name) DO UPDATE SET
				size = excluded.size, modtime = excluded.modtime, hash = excluded.hash`)
		if err != nil {
			fm.Close()
			fe.Close()
			t.Rollback()
			return err
		}
		tx, folderEnsureStmt, folderMtimeStmt, fileStmt, pending = t, fe, fm, fls, 0
		c.log.Debug("writer: tx opened")
		return nil
	}

	commitTx := func() {
		if tx == nil {
			return
		}
		committed := pending
		folderEnsureStmt.Close()
		folderMtimeStmt.Close()
		fileStmt.Close()
		if err := tx.Commit(); err != nil {
			c.log.Warn("cache: commit failed", "err", err)
			tx.Rollback()
			idCache = make(map[string]int64)
		}
		commitCount++
		c.log.Debug("writer: commit", "ops_in_tx", committed, "total_commits", commitCount)
		tx, folderEnsureStmt, folderMtimeStmt, fileStmt, pending = nil, nil, nil, nil, 0
	}

	ensureFolderID := func(folder string) (int64, error) {
		if id, ok := idCache[folder]; ok {
			return id, nil
		}
		var id int64
		if err := folderEnsureStmt.QueryRow(folder).Scan(&id); err != nil {
			return 0, err
		}
		idCache[folder] = id
		return id, nil
	}

	handleWrite := func(p *writePayload) {
		if tx == nil {
			if err := openTx(); err != nil {
				c.log.Warn("cache: begin failed", "err", err)
				return
			}
		}
		folder, name := splitFolderPath(p.path)
		folderID, err := ensureFolderID(folder)
		if err != nil {
			c.log.Warn("cache: resolve folder failed", "path", p.path, "err", err)
			return
		}
		if _, err := fileStmt.Exec(folderID, name, p.entry.Size, p.entry.ModTime.UnixNano(), p.entry.Hash[:]); err != nil {
			c.log.Warn("cache: insert failed", "path", p.path, "err", err)
			return
		}
		pending++
		if pending >= flushSize {
			commitTx()
		}
	}

	handleSetFolderMtime := func(p *folderMtimePayload) {
		if tx == nil {
			if err := openTx(); err != nil {
				c.log.Warn("cache: begin failed", "err", err)
				return
			}
		}
		var id int64
		if err := folderMtimeStmt.QueryRow(p.path, p.mtime.UnixNano()).Scan(&id); err != nil {
			c.log.Warn("cache: set folder mtime failed", "folder", p.path, "err", err)
			return
		}
		idCache[p.path] = id
		pending++
		if pending >= flushSize {
			commitTx()
		}
	}

	handle := func(op cacheOp) {
		switch {
		case op.write != nil:
			handleWrite(op.write)
		case op.setFolderMtime != nil:
			handleSetFolderMtime(op.setFolderMtime)
		case op.flush != nil:
			commitTx()
			op.flush.done <- struct{}{}
		case op.sweep != nil:
			commitTx()
			op.sweep.done <- c.runSweep(op.sweep.seenFolders, op.sweep.seenFiles, op.sweep.roots)
			idCache = make(map[string]int64)
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
			c.log.Debug("writer: shutdown signal received; draining")
			drained := 0
			for {
				select {
				case op := <-c.writes:
					handle(op)
					drained++
				default:
					commitTx()
					c.log.Debug("writer: exit", "drained_on_shutdown", drained, "total_commits", commitCount)
					return
				}
			}
		}
	}
}

// runSweep drops orphan folders (CASCADE drops files) and then orphan files within seen folders.
func (c *Cache) runSweep(seenFolders map[string]struct{}, seenFiles map[string]map[string]struct{}, roots []string) error {
	var totalSeenFiles int
	for _, names := range seenFiles {
		totalSeenFiles += len(names)
	}
	c.log.Debug("sweep: start", "seen_folders", len(seenFolders), "seen_files", totalSeenFiles, "roots", len(roots))

	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("begin sweep tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("CREATE TEMP TABLE sweep_seen_folders(path TEXT PRIMARY KEY)"); err != nil {
		return fmt.Errorf("create temp folders: %w", err)
	}
	if _, err := tx.Exec("CREATE TEMP TABLE sweep_seen_files(folder_path TEXT, name TEXT, PRIMARY KEY (folder_path, name))"); err != nil {
		return fmt.Errorf("create temp files: %w", err)
	}
	c.log.Debug("sweep: temp tables created")

	insF, err := tx.Prepare("INSERT INTO sweep_seen_folders(path) VALUES (?)")
	if err != nil {
		return fmt.Errorf("prepare seen folders insert: %w", err)
	}
	for p := range seenFolders {
		if _, err := insF.Exec(p); err != nil {
			insF.Close()
			return fmt.Errorf("insert seen folder: %w", err)
		}
	}
	insF.Close()

	insFL, err := tx.Prepare("INSERT INTO sweep_seen_files(folder_path, name) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("prepare seen files insert: %w", err)
	}
	for folder, names := range seenFiles {
		for name := range names {
			if _, err := insFL.Exec(folder, name); err != nil {
				insFL.Close()
				return fmt.Errorf("insert seen file: %w", err)
			}
		}
	}
	insFL.Close()
	c.log.Debug("sweep: seen sets loaded")

	// Count files that will be cascade-removed when their orphan folder is dropped.
	countCascade, err := tx.Prepare(`SELECT COUNT(*) FROM files
		WHERE folder_id IN (
			SELECT id FROM folders
			 WHERE (path = ?1 OR substr(path, 1, length(?1)+1) = ?1 || '/')
			   AND path NOT IN (SELECT path FROM sweep_seen_folders)
		)`)
	if err != nil {
		return fmt.Errorf("prepare cascade count: %w", err)
	}
	var cascadeFiles int64
	for _, root := range roots {
		var n int64
		if err := countCascade.QueryRow(root).Scan(&n); err != nil {
			countCascade.Close()
			return fmt.Errorf("count cascade files: %w", err)
		}
		cascadeFiles += n
	}
	countCascade.Close()
	c.log.Debug("sweep: cascade counted", "cascade_files", cascadeFiles)

	delFolders, err := tx.Prepare(`DELETE FROM folders
		WHERE (path = ?1 OR substr(path, 1, length(?1)+1) = ?1 || '/')
		  AND path NOT IN (SELECT path FROM sweep_seen_folders)`)
	if err != nil {
		return fmt.Errorf("prepare delete folders: %w", err)
	}
	var foldersDeleted int64
	for _, root := range roots {
		res, err := delFolders.Exec(root)
		if err != nil {
			delFolders.Close()
			return fmt.Errorf("delete folders: %w", err)
		}
		if n, err := res.RowsAffected(); err == nil {
			foldersDeleted += n
		}
	}
	delFolders.Close()
	c.log.Debug("sweep: folders deleted", "count", foldersDeleted)

	res, err := tx.Exec(`DELETE FROM files
		WHERE folder_id IN (SELECT id FROM folders WHERE path IN (SELECT path FROM sweep_seen_folders))
		  AND (folder_id, name) NOT IN (
		      SELECT d.id, sf.name FROM sweep_seen_files sf JOIN folders d ON d.path = sf.folder_path
		  )`)
	if err != nil {
		return fmt.Errorf("delete files: %w", err)
	}
	directFilesDeleted, _ := res.RowsAffected()
	c.log.Debug("sweep: direct files deleted", "count", directFilesDeleted)

	if _, err := tx.Exec("DROP TABLE sweep_seen_files"); err != nil {
		return fmt.Errorf("drop temp files: %w", err)
	}
	if _, err := tx.Exec("DROP TABLE sweep_seen_folders"); err != nil {
		return fmt.Errorf("drop temp folders: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	c.log.Debug("sweep: committed")
	c.log.Info("cache sweep complete",
		"folders_deleted", foldersDeleted,
		"files_deleted", cascadeFiles+directFilesDeleted,
	)
	return nil
}

// splitFolderPath splits an absolute file path into (folder, basename).
func splitFolderPath(p string) (folder, name string) {
	folder, name = filepath.Split(p)
	folder = strings.TrimRight(folder, string(filepath.Separator))
	if folder == "" {
		folder = string(filepath.Separator)
	}
	return folder, name
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
