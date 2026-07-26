package importer

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"local.dev/cine/database"
	"local.dev/cine/reader"
)

func TestBuildMetadata(t *testing.T) {
	ctx := context.Background()

	t.Run("appendSource records the file's own timestamp and size", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, reader.FileTitleBasics)
		require.NoError(t, os.WriteFile(path, []byte("0123456789"), 0o644))

		// The wave IMDb actually published title.basics in, on 25 July 2026.
		modified := time.Date(2026, 7, 25, 12, 35, 3, 0, time.UTC)
		require.NoError(t, os.Chtimes(path, modified, modified))

		sources, err := appendSource(nil, dir, reader.FileTitleBasics, 42)
		require.NoError(t, err)
		require.Len(t, sources, 1)
		assert.Equal(t, sourceRow{
			file:         reader.FileTitleBasics,
			lastModified: "2026-07-25T12:35:03Z",
			bytes:        10,
			rowsRead:     42,
		}, sources[0])
	})

	t.Run("a missing source file is an error, not a silent gap", func(t *testing.T) {
		_, err := appendSource(nil, t.TempDir(), reader.FileTitleBasics, 0)
		assert.ErrorContains(t, err, reader.FileTitleBasics)
	})

	t.Run("writeBuildMetadata fills both tables", func(t *testing.T) {
		db := openImportDB(t)
		sources := []sourceRow{
			{file: reader.FileTitleRatings, lastModified: "2026-07-25T00:33:47Z", bytes: 8578278, rowsRead: 1697944},
			{file: reader.FileTitleBasics, lastModified: "2026-07-25T12:35:03Z", bytes: 224630806, rowsRead: 12664477},
		}
		started := time.Date(2026, 7, 26, 1, 0, 0, 0, time.UTC)
		finished := started.Add(30 * time.Minute)

		require.NoError(t, inTx(ctx, db, func(tx *sql.Tx) error {
			return writeBuildMetadata(ctx, tx, sources, started, finished)
		}))

		t.Run("one build_sources row per file, keeping each wave apart", func(t *testing.T) {
			rows, err := db.QueryContext(ctx,
				"SELECT file, last_modified, bytes, rows_read FROM build_sources ORDER BY file")
			require.NoError(t, err)
			defer rows.Close()

			var got []sourceRow
			for rows.Next() {
				var s sourceRow
				require.NoError(t, rows.Scan(&s.file, &s.lastModified, &s.bytes, &s.rowsRead))
				got = append(got, s)
			}
			require.NoError(t, rows.Err())
			assert.ElementsMatch(t, sources, got)
		})

		t.Run("build_info holds the single build row", func(t *testing.T) {
			var (
				layer                 int64
				version               string
				startedAt, finishedAt string
			)
			require.NoError(t, db.QueryRowContext(ctx,
				"SELECT layer, cine_version, started_at, finished_at FROM build_info").
				Scan(&layer, &version, &startedAt, &finishedAt))
			assert.Equal(t, int64(layerFull), layer)
			assert.Equal(t, Version, version)
			assert.Equal(t, "2026-07-26T01:00:00Z", startedAt)
			assert.Equal(t, "2026-07-26T01:30:00Z", finishedAt)
		})
	})

	t.Run("a full build records all seven source files", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "cine.db")
		require.NoError(t, Import(ctx, out, gzipFixtures(t), log.New(io.Discard)))

		_, db, err := database.Open(ctx, out)
		require.NoError(t, err)
		t.Cleanup(func() { db.Close() })

		var files []string
		rows, err := db.QueryContext(ctx, "SELECT file FROM build_sources ORDER BY file")
		require.NoError(t, err)
		defer rows.Close()
		for rows.Next() {
			var file string
			require.NoError(t, rows.Scan(&file))
			files = append(files, file)
		}
		require.NoError(t, rows.Err())

		assert.ElementsMatch(t, []string{
			reader.FileTitleBasics, reader.FileTitleRatings, reader.FileNameBasics,
			reader.FileTitleEpisode, reader.FileTitleCrew, reader.FileTitlePrincipals,
			reader.FileTitleAkas,
		}, files)

		// rows_read is the only surviving record of title.ratings, which has no
		// table of its own once folded into titles.
		var ratingsRead int64
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT rows_read FROM build_sources WHERE file = ?", reader.FileTitleRatings).Scan(&ratingsRead))
		assert.Equal(t, int64(2), ratingsRead)
	})
}

// gzipFixtures copies testdata/imdb into a temporary folder as the gzipped files
// a real build reads, so Import can be exercised end to end.
func gzipFixtures(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range []string{
		reader.FileTitleBasics, reader.FileTitleRatings, reader.FileNameBasics,
		reader.FileTitleEpisode, reader.FileTitleCrew, reader.FileTitlePrincipals,
		reader.FileTitleAkas,
	} {
		plain, err := os.ReadFile(filepath.Join("testdata", "imdb", strings.TrimSuffix(name, ".gz")))
		require.NoError(t, err)

		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		_, err = gz.Write(plain)
		require.NoError(t, err)
		require.NoError(t, gz.Close())

		require.NoError(t, os.WriteFile(filepath.Join(dir, name), buf.Bytes(), 0o644))
	}
	return dir
}
