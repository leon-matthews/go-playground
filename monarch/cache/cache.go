// Package cache stores mediainfo results in a per-folder JSON file.
package cache

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go-playground/monarch/mediainfo"
)

// FileName is the cache file written into each scanned folder.
const FileName = ".monarch-cache.json"

// schema versions the on-disk format; bump it when entry fields change.
const schema = 1

// ttl is how long an entry stays valid after it was cached.
const ttl = 24 * time.Hour

// entry is a single cached file's metadata and staleness markers.
type entry struct {
	Size     int64            `json:"size"`
	ModTime  time.Time        `json:"mod_time"`
	CachedAt time.Time        `json:"cached_at"`
	Media    *mediainfo.Media `json:"media"`
}

// document wraps the entries with the versions they are valid for.
type document struct {
	Schema           int              `json:"schema"`
	MediainfoVersion string           `json:"mediainfo_version"`
	Entries          map[string]entry `json:"entries"`
}

// Cache is a per-folder media metadata cache backed by a JSON file.
type Cache struct {
	dir     string
	version string
	now     time.Time
	loaded  map[string]entry
	live    map[string]entry
	mu      sync.Mutex
}

// Load reads the cache file from dir, dropping entries that are expired or
// were written for a different mediainfo or schema version.
//
// A missing, unreadable, or stale cache file yields an empty cache.
func Load(dir, mediainfoVersion string) (*Cache, error) {
	now := time.Now()
	c := &Cache{
		dir:     dir,
		version: mediainfoVersion,
		now:     now,
		loaded:  map[string]entry{},
		live:    map[string]entry{},
	}

	data, err := os.ReadFile(filepath.Join(dir, FileName))
	if errors.Is(err, fs.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return c, err
	}

	var doc document
	if err := json.Unmarshal(data, &doc); err != nil {
		return c, nil
	}
	if doc.Schema != schema || doc.MediainfoVersion != mediainfoVersion {
		return c, nil
	}

	for name, e := range doc.Entries {
		if now.Sub(e.CachedAt) <= ttl {
			c.loaded[name] = e
		}
	}
	return c, nil
}

// Get returns the cached media for name when its size and modification time
// match, marking the entry to be kept when the cache is next saved.
func (c *Cache) Get(name string, size int64, modTime time.Time) (*mediainfo.Media, bool) {
	e, ok := c.loaded[name]
	if !ok || e.Size != size || !e.ModTime.Equal(modTime) {
		return nil, false
	}
	c.mu.Lock()
	c.live[name] = e
	c.mu.Unlock()
	return e.Media, true
}

// Put records freshly read media for name, replacing any existing entry.
func (c *Cache) Put(name string, size int64, modTime time.Time, media *mediainfo.Media) {
	c.mu.Lock()
	c.live[name] = entry{Size: size, ModTime: modTime, CachedAt: c.now, Media: media}
	c.mu.Unlock()
}

// Save writes the entries gathered since Load back to the cache file.
//
// Only files seen during this run are kept, so deleted files fall out.
func (c *Cache) Save() error {
	doc := document{
		Schema:           schema,
		MediainfoVersion: c.version,
		Entries:          c.live,
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.dir, FileName), data, 0o644)
}
