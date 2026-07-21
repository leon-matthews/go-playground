package mimicry

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

	_ "modernc.org/sqlite" // registers the "sqlite" database/sql driver
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

// sweepDeleteChunk caps the number of names in a single `DELETE FROM files ... IN (...)`
// during sweep, staying well clear of SQLite's parameter limit.
const sweepDeleteChunk = 500

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

// OpenCache opens or creates the SQLite cache at path; rebuilds the schema if it's stale or absent.
func OpenCache(path string, log *slog.Logger) (*Cache, error) {
	return openCacheWithTimeout(path, openTimeout, log)
}

// openCacheWithTimeout is OpenCache with an explicit ping timeout for tests.
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

	_, statErr := os.Stat(path)
	freshFile := errors.Is(statErr, os.ErrNotExist)

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

	if err := ensureSchema(db, log); err != nil {
		db.Close()
		return &Cache{log: log}, err
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
	if freshFile {
		log.Info("new cache file created", "path", path)
	} else {
		log.Debug("cache opened", "path", path)
	}
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
func (c *Cache) Sweep(folderScans []FolderInfo, roots []string) error {
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
func buildSweepSets(folderInfos []FolderInfo) (map[string]struct{}, map[string]map[string]struct{}) {
	seenFolders := make(map[string]struct{}, len(folderInfos))
	seenFiles := make(map[string]map[string]struct{}, len(folderInfos))
	for _, f := range folderInfos {
		seenFolders[f.Path] = struct{}{}
		names := make(map[string]struct{}, len(f.Children))
		for _, name := range f.Children {
			names[name] = struct{}{}
		}
		seenFiles[f.Path] = names
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

// AllFiles returns every cached file as a FileInfo, reconstructing full paths from the folders
// join and deriving each extension from the name. Order is unspecified.
//
// This is a cold, one-shot read (the report path), so it runs a plain query rather than a
// prepared statement kept open on the Cache.
func (c *Cache) AllFiles() ([]FileInfo, error) {
	if c == nil || c.db == nil {
		return nil, nil
	}
	rows, err := c.db.Query(`SELECT d.path, f.name, f.size, f.modtime, f.hash
		FROM files f JOIN folders d ON d.id = f.folder_id`)
	if err != nil {
		return nil, fmt.Errorf("query all files: %w", err)
	}
	defer rows.Close()

	var out []FileInfo
	for rows.Next() {
		var (
			folder  string
			name    string
			size    int64
			modtime int64
			digest  []byte
		)
		if err := rows.Scan(&folder, &name, &size, &modtime, &digest); err != nil {
			return nil, fmt.Errorf("scan file row: %w", err)
		}
		if len(digest) != sha256.Size {
			c.log.Warn("cache: unexpected hash length; skipping", "folder", folder, "name", name, "len", len(digest))
			continue
		}
		fi := FileInfo{
			Path:      filepath.Join(folder, name),
			Size:      size,
			ModTime:   time.Unix(0, modtime).UTC(),
			Extension: filepath.Ext(name),
		}
		copy(fi.Hash[:], digest)
		out = append(out, fi)
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
		tx               *sql.Tx
		folderEnsureStmt *sql.Stmt
		folderMtimeStmt  *sql.Stmt
		fileStmt         *sql.Stmt
		pending          int
		idCache          = make(map[string]int64)
		commitCount      int
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

// runSweep drops orphan folders (CASCADE drops their files), then drops files inside seen
// folders that no longer exist on disk. The seen-folders set lives in a temp table so the
// folder DELETE can join against it; the seen-files diff happens Go-side per folder.
func (c *Cache) runSweep(seenFolders map[string]struct{}, seenFiles map[string]map[string]struct{}, roots []string) error {
	start := time.Now()

	tx, err := c.db.Begin()
	if err != nil {
		return fmt.Errorf("begin sweep tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("CREATE TEMP TABLE sweep_seen_folders(path TEXT PRIMARY KEY)"); err != nil {
		return fmt.Errorf("create temp folders: %w", err)
	}
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

	filesDeleted, err := sweepOrphanFiles(tx, seenFiles)
	if err != nil {
		return err
	}

	if _, err := tx.Exec("DROP TABLE sweep_seen_folders"); err != nil {
		return fmt.Errorf("drop temp folders: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	c.log.Info(
		"cache sweep complete",
		"folders_deleted", foldersDeleted,
		"files_deleted", filesDeleted,
		slog.Duration("elapsed", time.Since(start)),
	)
	return nil
}

// sweepOrphanFiles deletes file rows in each seen folder whose name no longer appears in the
// folder's seen-names set. For each folder it fetches current names, computes the diff in Go,
// and batch-deletes the orphans. Returns the total rows deleted.
func sweepOrphanFiles(tx *sql.Tx, seenFiles map[string]map[string]struct{}) (int64, error) {
	folderIDStmt, err := tx.Prepare("SELECT id FROM folders WHERE path = ?")
	if err != nil {
		return 0, fmt.Errorf("prepare folder id lookup: %w", err)
	}
	defer folderIDStmt.Close()

	namesStmt, err := tx.Prepare("SELECT name FROM files WHERE folder_id = ?")
	if err != nil {
		return 0, fmt.Errorf("prepare names lookup: %w", err)
	}
	defer namesStmt.Close()

	var total int64
	for folder, seenNames := range seenFiles {
		var folderID int64
		switch err := folderIDStmt.QueryRow(folder).Scan(&folderID); {
		case errors.Is(err, sql.ErrNoRows):
			continue
		case err != nil:
			return total, fmt.Errorf("lookup folder id: %w", err)
		}

		orphans, err := collectOrphanNames(namesStmt, folderID, seenNames)
		if err != nil {
			return total, err
		}
		if len(orphans) == 0 {
			continue
		}
		n, err := deleteOrphanFiles(tx, folderID, orphans)
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}

// collectOrphanNames returns the names currently in `files` for folderID that aren't in seen.
func collectOrphanNames(stmt *sql.Stmt, folderID int64, seen map[string]struct{}) ([]string, error) {
	rows, err := stmt.Query(folderID)
	if err != nil {
		return nil, fmt.Errorf("query names: %w", err)
	}
	defer rows.Close()
	var orphans []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan name: %w", err)
		}
		if _, ok := seen[name]; !ok {
			orphans = append(orphans, name)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate names: %w", err)
	}
	return orphans, nil
}

// deleteOrphanFiles removes (folderID, name) rows in chunks of sweepDeleteChunk.
func deleteOrphanFiles(tx *sql.Tx, folderID int64, names []string) (int64, error) {
	var total int64
	for begin := 0; begin < len(names); begin += sweepDeleteChunk {
		end := begin + sweepDeleteChunk
		if end > len(names) {
			end = len(names)
		}
		chunk := names[begin:end]
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(chunk)), ",")
		args := make([]any, 0, len(chunk)+1)
		args = append(args, folderID)
		for _, n := range chunk {
			args = append(args, n)
		}
		res, err := tx.Exec("DELETE FROM files WHERE folder_id = ? AND name IN ("+placeholders+")", args...)
		if err != nil {
			return total, fmt.Errorf("delete orphan files: %w", err)
		}
		if n, err := res.RowsAffected(); err == nil {
			total += n
		}
	}
	return total, nil
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
