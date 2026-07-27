package database_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"local.dev/cine/database"
	"local.dev/cine/database/sqlite"
)

// open builds a fresh database in a per-test temporary folder.
func open(t testing.TB) (*sqlite.Queries, *sql.DB) {
	t.Helper()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "test.db")
	queries, db, err := database.Open(ctx, path)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return queries, db
}

func TestDatabase(t *testing.T) {
	t.Run("schema creates every table", func(t *testing.T) {
		ctx := context.Background()
		_, db := open(t)

		rows, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type = 'table'")
		require.NoError(t, err)
		defer rows.Close()

		var got []string
		for rows.Next() {
			var name string
			require.NoError(t, rows.Scan(&name))
			got = append(got, name)
		}
		require.NoError(t, rows.Err())

		want := []string{
			"build_info", "build_sources",
			"titles_types", "genres", "titles",
			"professions", "names", "names_primary_professions", "names_known_for_titles",
			"episodes", "titles_credit_names",
			"principals_categories", "principals_jobs", "principals",
			"regions", "languages", "akas_types", "attributes", "akas", "akas_carry_attributes",
		}
		assert.ElementsMatch(t, want, got)
	})

	t.Run("schema.sql sets the user_version SchemaVersion promises", func(t *testing.T) {
		ctx := context.Background()
		_, db := open(t)

		var version int
		require.NoError(t, db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version))
		assert.Equal(t, database.SchemaVersion, version)
	})

	t.Run("build_info holds at most one row", func(t *testing.T) {
		ctx := context.Background()
		_, db := open(t)

		const insert = `INSERT INTO build_info (id, cine_version, started_at, finished_at)
			VALUES (?, '0.1.0', '2026-07-26T00:00:00Z', '2026-07-26T00:30:00Z')`
		_, err := db.ExecContext(ctx, insert, 1)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx, insert, 2)
		assert.ErrorContains(t, err, "CHECK constraint failed")
	})

	t.Run("only importer-generated references are declared", func(t *testing.T) {
		ctx := context.Background()
		_, db := open(t)

		rows, err := db.QueryContext(ctx, `
			SELECT m.name || '.' || f."from" || ' -> ' || f."table"
			FROM sqlite_master m
			JOIN pragma_foreign_key_list(m.name) f
			WHERE m.type = 'table'`)
		require.NoError(t, err)
		defer rows.Close()

		var got []string
		for rows.Next() {
			var ref string
			require.NoError(t, rows.Scan(&ref))
			got = append(got, ref)
		}
		require.NoError(t, rows.Err())

		// titles and names must never appear as a parent: IMDb's own files disagree.
		want := []string{
			"titles.title_type -> titles_types",
			"names_primary_professions.profession_id -> professions",
			"principals.category -> principals_categories",
			"principals.job -> principals_jobs",
			"akas.region -> regions",
			"akas.language -> languages",
			"akas_carry_attributes.attribute_id -> attributes",
			"akas_carry_attributes.title_id -> akas", // composite key, one row per column
			"akas_carry_attributes.ordering -> akas",
		}
		assert.ElementsMatch(t, want, got)
	})

	t.Run("pragmas are live on the connection", func(t *testing.T) {
		ctx := context.Background()
		_, db := open(t)

		cases := []struct {
			pragma string
			want   string
		}{
			{"journal_mode", "off"},
			{"synchronous", "0"}, // OFF
			{"locking_mode", "exclusive"},
			{"cache_size", "-65536"},
			{"temp_store", "2"}, // MEMORY
		}
		for _, tc := range cases {
			t.Run(tc.pragma, func(t *testing.T) {
				var got string
				require.NoError(t, db.QueryRowContext(ctx, "PRAGMA "+tc.pragma).Scan(&got))
				assert.Equal(t, tc.want, got)
			})
		}
	})

	t.Run("title round-trips through the generated queries", func(t *testing.T) {
		ctx := context.Background()
		queries, _ := open(t)

		id, err := queries.UpsertGenre(ctx, sqlite.UpsertGenreParams{ID: 8, Name: "Drama"})
		require.NoError(t, err)
		assert.Equal(t, int64(8), id)

		const actionSciFi = (1 << 0) | (1 << 21) // Action + Sci-Fi bits
		insert := sqlite.InsertTitleParams{
			ID:             133093,
			TitleType:      1,
			PrimaryTitle:   "The Matrix",
			OriginalTitle:  sql.NullString{}, // NULL: same as primary_title
			StartYear:      sql.NullInt64{Int64: 1999, Valid: true},
			RuntimeMinutes: sql.NullInt64{Int64: 136, Valid: true},
			Genres:         actionSciFi,
			AverageRating:  sql.NullInt64{Int64: 87, Valid: true}, // 8.7
			NumVotes:       sql.NullInt64{Int64: 2000000, Valid: true},
		}
		require.NoError(t, queries.InsertTitle(ctx, insert))

		got, err := queries.GetTitle(ctx, 133093)
		require.NoError(t, err)
		assert.Equal(t, "The Matrix", got.PrimaryTitle)
		assert.False(t, got.OriginalTitle.Valid) // stored NULL, not the empty string
		assert.Equal(t, int64(1999), got.StartYear.Int64)
		assert.Equal(t, int64(actionSciFi), got.Genres)
		assert.Equal(t, int64(87), got.AverageRating.Int64)
	})
}
