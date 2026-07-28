package importer

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"local.dev/cine/database"
	"local.dev/cine/logging"
	"local.dev/cine/reader"
)

func TestBuildMetadata(t *testing.T) {
	ctx := context.Background()

	t.Run("appendSource records the file's own timestamp and size", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, reader.FileTitleBasics)
		require.NoError(t, os.WriteFile(path, []byte("0123456789"), 0o644))

		// The timestamp IMDb's own title.basics carried on 25 July 2026.
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

		options := BuildOptions{Rated: true, NotAdult: true, People: true}
		require.NoError(t, inTx(ctx, db, func(tx *sql.Tx) error {
			return writeBuildMetadata(ctx, tx, sources, options, started, finished)
		}))

		t.Run("one build_sources row per file, keeping each timestamp apart", func(t *testing.T) {
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
				version               string
				startedAt, finishedAt string
			)
			require.NoError(t, db.QueryRowContext(ctx,
				"SELECT cine_version, started_at, finished_at FROM build_info").
				Scan(&version, &startedAt, &finishedAt))
			assert.Equal(t, Version, version)
			assert.Equal(t, "2026-07-26T01:00:00Z", startedAt)
			assert.Equal(t, "2026-07-26T01:30:00Z", finishedAt)
		})

		t.Run("build_info records every option the build was given", func(t *testing.T) {
			var rated, notAdult, people int64
			require.NoError(t, db.QueryRowContext(ctx,
				"SELECT filter_rated, filter_not_adult, has_people FROM build_info").
				Scan(&rated, &notAdult, &people))
			assert.Equal(t, int64(1), rated)
			assert.Equal(t, int64(1), notAdult)
			assert.Equal(t, int64(1), people)
		})
	})

	t.Run("a filtered build keeps only the allowed titles and their rows", func(t *testing.T) {
		// Only tt0000001 and tt0133093 are rated, and no episode's parent is, so the
		// rated rule leaves the two films and drops everything keyed to any other id.
		out := filepath.Join(t.TempDir(), "cine.db")
		options := BuildOptions{Rated: true, People: true}
		require.NoError(t, Import(ctx, out, gzipFixtures(t), options, logging.Discard()))

		_, db, err := database.Open(ctx, out, true)
		require.NoError(t, err)
		t.Cleanup(func() { db.Close() })

		count := func(t *testing.T, table string) int {
			t.Helper()
			var n int
			require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&n))
			return n
		}

		t.Run("titles keyed to a refused id are gone from every table", func(t *testing.T) {
			assert.Equal(t, 2, count(t, "titles"), "tt9999999 is unrated")
			assert.Zero(t, count(t, "episodes"), "neither parent series is rated")
			assert.Equal(t, 3, count(t, "titles_credit_names"), "tt0000002's writer dropped")
			assert.Equal(t, 2, count(t, "principals"))
			assert.Equal(t, 2, count(t, "akas"))
		})

		t.Run("names are not filtered, being keyed to people not titles", func(t *testing.T) {
			assert.Equal(t, 3, count(t, "names"))
		})

		t.Run("build_sources still counts source rows, not written ones", func(t *testing.T) {
			var read int64
			require.NoError(t, db.QueryRowContext(ctx,
				"SELECT rows_read FROM build_sources WHERE file = ?", reader.FileTitleBasics).Scan(&read))
			assert.Equal(t, int64(3), read, "provenance describes the file, not the build")
		})

		t.Run("build_info records the rule that ran", func(t *testing.T) {
			var rated, notAdult int64
			require.NoError(t, db.QueryRowContext(ctx,
				"SELECT filter_rated, filter_not_adult FROM build_info").Scan(&rated, &notAdult))
			assert.Equal(t, int64(1), rated)
			assert.Equal(t, int64(0), notAdult)
		})

		t.Run("known-for rows point only at titles the filter kept", func(t *testing.T) {
			var dangling int
			require.NoError(t, db.QueryRowContext(ctx, `SELECT count(*) FROM names_known_for_titles k
				LEFT JOIN titles t ON t.id = k.title_id WHERE t.id IS NULL`).Scan(&dangling))
			assert.Zero(t, dangling, "the names pass takes the filter for this junction alone")
		})

		t.Run("generated lookup ids all still resolve", func(t *testing.T) {
			rows, err := db.QueryContext(ctx, "PRAGMA foreign_key_check")
			require.NoError(t, err)
			defer rows.Close()
			assert.False(t, rows.Next(), "a filtered build must leave no dangling lookup id")
			require.NoError(t, rows.Err())
		})
	})

	t.Run("a full build records all seven source files", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "cine.db")
		require.NoError(t, Import(ctx, out, gzipFixtures(t), BuildOptions{People: true}, logging.Discard()))

		_, db, err := database.Open(ctx, out, true)
		require.NoError(t, err)
		t.Cleanup(func() { db.Close() })

		assert.ElementsMatch(t, []string{
			reader.FileTitleBasics, reader.FileTitleRatings, reader.FileNameBasics,
			reader.FileTitleEpisode, reader.FileTitleCrew, reader.FileTitlePrincipals,
			reader.FileTitleAkas,
		}, sourceFiles(ctx, t, db))

		// rows_read is the only surviving record of title.ratings, which has no
		// table of its own once folded into titles.
		var ratingsRead int64
		require.NoError(t, db.QueryRowContext(ctx,
			"SELECT rows_read FROM build_sources WHERE file = ?", reader.FileTitleRatings).Scan(&ratingsRead))
		assert.Equal(t, int64(2), ratingsRead)
	})

	t.Run("a build without people reads only the titles files", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "cine.db")
		require.NoError(t, Import(ctx, out, gzipFixtures(t), BuildOptions{}, logging.Discard()))

		_, db, err := database.Open(ctx, out, false)
		require.NoError(t, err)
		t.Cleanup(func() { db.Close() })

		count := func(t *testing.T, table string) int {
			t.Helper()
			var n int
			require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM "+table).Scan(&n))
			return n
		}

		t.Run("the three people files are absent from build_sources", func(t *testing.T) {
			// An absent row is how a table nobody imported is told apart from an empty one.
			assert.ElementsMatch(t, []string{
				reader.FileTitleBasics, reader.FileTitleRatings,
				reader.FileTitleEpisode, reader.FileTitleAkas,
			}, sourceFiles(ctx, t, db))
		})

		t.Run("the titles tables are populated as usual", func(t *testing.T) {
			assert.Equal(t, 3, count(t, "titles"))
			assert.Equal(t, 2, count(t, "episodes"))
			assert.Equal(t, 3, count(t, "akas"))
		})

		t.Run("no people table was created at all", func(t *testing.T) {
			for _, table := range []string{
				"names", "names_primary_professions", "names_known_for_titles",
				"names_profession", "titles_credit_names", "principals",
				"principals_category", "principals_job",
			} {
				var name string
				err := db.QueryRowContext(ctx,
					"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&name)
				assert.ErrorIs(t, err, sql.ErrNoRows, table)
			}
		})

		t.Run("build_info records that people were not imported", func(t *testing.T) {
			var people int64
			require.NoError(t, db.QueryRowContext(ctx,
				"SELECT has_people FROM build_info").Scan(&people))
			assert.Equal(t, int64(0), people)
		})
	})
}

// sourceFiles reads the file column of every build_sources row.
func sourceFiles(ctx context.Context, t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.QueryContext(ctx, "SELECT file FROM build_sources ORDER BY file")
	require.NoError(t, err)
	defer rows.Close()

	var files []string
	for rows.Next() {
		var file string
		require.NoError(t, rows.Scan(&file))
		files = append(files, file)
	}
	require.NoError(t, rows.Err())
	return files
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
