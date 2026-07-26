package importer

import (
	"context"
	"database/sql"
	"io"
	"math"

	"local.dev/cine/reader"
)

// titleColumns are the titles columns in the order bindTitleRow writes them.
var titleColumns = []string{
	"id", "title_type", "primary_title", "original_title", "is_adult",
	"start_year", "end_year", "runtime_minutes", "genres", "average_rating", "num_votes",
}

// rating holds the two title.ratings values for one title.
type rating struct {
	average int64 // stored in tenths, so 5.7 becomes 57
	votes   int64
}

// loadRatings reads title.ratings into a map keyed by integer title id, ready to
// join onto titles during their pass. It also reports how many source rows it
// read, which is the only record of them: ratings have no table of their own, and
// any whose title is absent from title.basics are dropped by the join.
func loadRatings(r io.Reader) (map[int64]rating, int64, error) {
	ratings := make(map[int64]rating)
	var read int64
	for rec, err := range reader.ReadTitleRatings(r) {
		if err != nil {
			return nil, 0, err
		}
		read++
		id, err := parseID(rec.Tconst)
		if err != nil {
			return nil, 0, err
		}
		ratings[id] = rating{
			average: int64(math.Round(rec.AverageRating * 10)),
			votes:   int64(rec.NumVotes),
		}
	}
	return ratings, read, nil
}

// titleLookups holds the interners populated while reading title.basics.
type titleLookups struct {
	titleType *interner
	genre     *interner
}

// importTitles streams title.basics into the titles table, joining ratings and
// interning the title_type and genre lookups. Titles the filter refuses are not
// written. It returns the number of titles written; the caller writes the
// returned lookups to their tables once the pass completes.
func importTitles(ctx context.Context, tx *sql.Tx, basics io.Reader, ratings map[int64]rating, filter titleFilter) (counts, *titleLookups, error) {
	lookups := &titleLookups{titleType: newInterner(), genre: newInterner()}
	inserter, err := newBatchInserter(ctx, tx, "titles", titleColumns, bindTitleRow)
	if err != nil {
		return counts{}, nil, err
	}
	var read int64
	for basic, err := range reader.ReadTitleBasics(basics) {
		if err != nil {
			return counts{}, nil, err
		}
		read++
		id, err := parseID(basic.Tconst)
		if err != nil {
			return counts{}, nil, err
		}
		// Refusing before the row is built keeps dropped values out of the interners.
		if !filter.allows(id) {
			continue
		}
		if err := inserter.Add(ctx, buildTitleRow(basic, id, ratings, lookups)); err != nil {
			return counts{}, nil, err
		}
	}
	if err := inserter.Flush(ctx); err != nil {
		return counts{}, nil, err
	}
	return counts{read: read, written: inserter.Added()}, lookups, nil
}

// titleRow holds one titles row's values in column order; a nil field is stored
// as SQL NULL.
type titleRow struct {
	id            int64
	titleType     int64
	primaryTitle  string
	originalTitle any
	isAdult       int64
	startYear     any
	endYear       any
	runtime       any
	genres        int64
	averageRating any
	numVotes      any
}

// buildTitleRow transforms a reader record into a titles row, interning its
// title type and genres and joining any rating.
func buildTitleRow(b reader.TitleBasics, id int64, ratings map[int64]rating, lookups *titleLookups) titleRow {
	var genres int64
	for _, name := range b.Genres {
		genres |= lookups.genre.bit(name)
	}

	row := titleRow{
		id:            id,
		titleType:     lookups.titleType.id(b.TitleType),
		primaryTitle:  b.PrimaryTitle,
		originalTitle: droppedIfEqual(b.OriginalTitle, b.PrimaryTitle),
		isAdult:       boolToInt(b.IsAdult),
		startYear:     nullableInt(b.StartYear),
		endYear:       nullableInt(b.EndYear),
		runtime:       nullableInt(b.RuntimeMinutes),
		genres:        genres,
	}
	if r, ok := ratings[id]; ok {
		row.averageRating = r.average
		row.numVotes = r.votes
	}
	return row
}

// bindTitleRow appends a titles row's values in titleColumns order.
func bindTitleRow(args []any, r titleRow) []any {
	return append(args,
		r.id, r.titleType, r.primaryTitle, r.originalTitle, r.isAdult,
		r.startYear, r.endYear, r.runtime, r.genres, r.averageRating, r.numVotes)
}

// nullableInt maps the reader's Missing sentinel to nil, so an absent optional
// integer is stored as SQL NULL.
func nullableInt(n int) any {
	if n == reader.Missing {
		return nil
	}
	return int64(n)
}

// nullableStr maps the reader's empty string to nil, so an absent optional
// string is stored as SQL NULL.
func nullableStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// droppedIfEqual returns nil when value equals other, so a redundant
// original_title is stored as NULL and reconstructed with coalesce on read.
func droppedIfEqual(value, other string) any {
	if value == other {
		return nil
	}
	return value
}

// boolToInt maps a boolean to the 0 or 1 SQLite stores.
func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
