package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-playground/monarch/mediainfo"
)

func TestCache(t *testing.T) {
	const version = "v26.01"
	const size = int64(483210)
	mtime := time.Date(2025, 8, 31, 11, 25, 21, 0, time.UTC)

	t.Run("miss on empty cache", func(t *testing.T) {
		c, err := Load(t.TempDir(), version)
		require.NoError(t, err)

		_, ok := c.Get("cow.mp4", size, mtime)
		assert.False(t, ok)
	})

	t.Run("hit on matching entry", func(t *testing.T) {
		dir := t.TempDir()
		writeCache(t, dir, freshDoc(version, mtime, size))

		c, err := Load(dir, version)
		require.NoError(t, err)

		got, ok := c.Get("cow.mp4", size, mtime)
		require.True(t, ok)
		assert.Equal(t, sampleMedia(), got)
	})

	t.Run("miss when size differs", func(t *testing.T) {
		dir := t.TempDir()
		writeCache(t, dir, freshDoc(version, mtime, size))

		c, err := Load(dir, version)
		require.NoError(t, err)

		_, ok := c.Get("cow.mp4", size+1, mtime)
		assert.False(t, ok)
	})

	t.Run("miss when mtime differs", func(t *testing.T) {
		dir := t.TempDir()
		writeCache(t, dir, freshDoc(version, mtime, size))

		c, err := Load(dir, version)
		require.NoError(t, err)

		_, ok := c.Get("cow.mp4", size, mtime.Add(time.Second))
		assert.False(t, ok)
	})

	t.Run("expired entry is pruned", func(t *testing.T) {
		dir := t.TempDir()
		doc := freshDoc(version, mtime, size)
		doc.Entries["cow.mp4"] = entry{
			Size: size, ModTime: mtime, CachedAt: time.Now().Add(-ttl - time.Minute), Media: sampleMedia(),
		}
		writeCache(t, dir, doc)

		c, err := Load(dir, version)
		require.NoError(t, err)

		_, ok := c.Get("cow.mp4", size, mtime)
		assert.False(t, ok)
	})

	t.Run("mediainfo version mismatch is ignored", func(t *testing.T) {
		dir := t.TempDir()
		writeCache(t, dir, freshDoc("v1.00", mtime, size))

		c, err := Load(dir, version)
		require.NoError(t, err)

		_, ok := c.Get("cow.mp4", size, mtime)
		assert.False(t, ok)
	})

	t.Run("schema mismatch is ignored", func(t *testing.T) {
		dir := t.TempDir()
		doc := freshDoc(version, mtime, size)
		doc.Schema = schema + 1
		writeCache(t, dir, doc)

		c, err := Load(dir, version)
		require.NoError(t, err)

		_, ok := c.Get("cow.mp4", size, mtime)
		assert.False(t, ok)
	})

	t.Run("corrupt file is ignored", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, FileName), []byte("{not json"), 0o644))

		c, err := Load(dir, version)
		require.NoError(t, err)

		_, ok := c.Get("cow.mp4", size, mtime)
		assert.False(t, ok)
	})

	t.Run("put then save round-trips", func(t *testing.T) {
		dir := t.TempDir()
		c, err := Load(dir, version)
		require.NoError(t, err)

		c.Put("cow.mp4", size, mtime, sampleMedia())
		require.NoError(t, c.Save())

		reloaded, err := Load(dir, version)
		require.NoError(t, err)

		got, ok := reloaded.Get("cow.mp4", size, mtime)
		require.True(t, ok)
		assert.Equal(t, sampleMedia(), got)

		doc := readDoc(t, dir)
		assert.WithinDuration(t, time.Now(), doc.Entries["cow.mp4"].CachedAt, time.Minute)
	})

	t.Run("save keeps only files seen this run", func(t *testing.T) {
		dir := t.TempDir()
		writeCache(t, dir, document{
			Schema:           schema,
			MediainfoVersion: version,
			Entries: map[string]entry{
				"seen.mp4":     {Size: size, ModTime: mtime, CachedAt: time.Now().Add(-time.Hour), Media: sampleMedia()},
				"vanished.mp4": {Size: size, ModTime: mtime, CachedAt: time.Now().Add(-time.Hour), Media: sampleMedia()},
			},
		})

		c, err := Load(dir, version)
		require.NoError(t, err)

		_, ok := c.Get("seen.mp4", size, mtime)
		require.True(t, ok)
		require.NoError(t, c.Save())

		doc := readDoc(t, dir)
		assert.Contains(t, doc.Entries, "seen.mp4")
		assert.NotContains(t, doc.Entries, "vanished.mp4")
	})

	t.Run("hit preserves original cached time", func(t *testing.T) {
		dir := t.TempDir()
		cachedAt := time.Now().Add(-12 * time.Hour)
		writeCache(t, dir, document{
			Schema:           schema,
			MediainfoVersion: version,
			Entries: map[string]entry{
				"cow.mp4": {Size: size, ModTime: mtime, CachedAt: cachedAt, Media: sampleMedia()},
			},
		})

		c, err := Load(dir, version)
		require.NoError(t, err)

		_, ok := c.Get("cow.mp4", size, mtime)
		require.True(t, ok)
		require.NoError(t, c.Save())

		doc := readDoc(t, dir)
		assert.WithinDuration(t, cachedAt, doc.Entries["cow.mp4"].CachedAt, time.Second)
	})
}

// freshDoc builds a valid document holding a single, recently cached entry.
func freshDoc(version string, mtime time.Time, size int64) document {
	return document{
		Schema:           schema,
		MediainfoVersion: version,
		Entries: map[string]entry{
			"cow.mp4": {Size: size, ModTime: mtime, CachedAt: time.Now().Add(-time.Hour), Media: sampleMedia()},
		},
	}
}

// sampleMedia returns a Media value for use as cached content.
func sampleMedia() *mediainfo.Media {
	return &mediainfo.Media{
		Name:           "cow.mp4",
		Size:           483210,
		Format:         "MPEG-4",
		OverallBitrate: 962570,
		Duration:       mediainfo.Duration(4 * time.Second),
		Video:          []mediainfo.VideoTrack{{Format: "HEVC", Bitrate: 819440, Width: 480, Height: 848}},
		Audio:          []mediainfo.AudioTrack{{Format: "AAC", Bitrate: 132300, Channels: 2}},
		Text:           []mediainfo.TextTrack{},
	}
}

// writeCache marshals doc to the cache file in dir.
func writeCache(t *testing.T, dir string, doc document) {
	t.Helper()
	data, err := json.Marshal(doc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, FileName), data, 0o644))
}

// readDoc reads the cache file from dir back into a document.
func readDoc(t *testing.T, dir string) document {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, FileName))
	require.NoError(t, err)
	var doc document
	require.NoError(t, json.Unmarshal(data, &doc))
	return doc
}
