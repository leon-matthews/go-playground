package importer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"

	"local.dev/cine/reader"
)

// principalColumns are the principals columns in the order bindPrincipalRow
// writes them.
var principalColumns = []string{"title_id", "ordering", "name_id", "category", "job", "characters"}

// principalLookups holds the interners populated while reading title.principals.
type principalLookups struct {
	category *interner
	job      *interner
}

// importPrincipals streams title.principals into the principals table, interning
// the category and job lookups. The caller writes them to their tables once the
// pass completes.
func importPrincipals(ctx context.Context, tx *sql.Tx, principals io.Reader) (*principalLookups, error) {
	lookups := &principalLookups{category: newInterner(), job: newInterner()}
	inserter, err := newBatchInserter(ctx, tx, "principals", principalColumns, bindPrincipalRow)
	if err != nil {
		return nil, err
	}
	for record, err := range reader.ReadTitlePrincipals(principals) {
		if err != nil {
			return nil, err
		}
		row, err := buildPrincipalRow(record, lookups)
		if err != nil {
			return nil, err
		}
		if err := inserter.Add(ctx, row); err != nil {
			return nil, err
		}
	}
	if err := inserter.Flush(ctx); err != nil {
		return nil, err
	}
	return lookups, nil
}

// principalRow holds one principals row's values in column order; a nil field is
// stored as SQL NULL.
type principalRow struct {
	titleID    int64
	ordering   int64
	nameID     int64
	category   int64
	job        any
	characters any
}

// buildPrincipalRow transforms a reader record into a principals row, interning
// its category and job and re-encoding its characters as JSON text.
func buildPrincipalRow(p reader.TitlePrincipals, lookups *principalLookups) (principalRow, error) {
	titleID, err := parseID(p.Tconst)
	if err != nil {
		return principalRow{}, err
	}
	nameID, err := parseID(p.Nconst)
	if err != nil {
		return principalRow{}, err
	}
	characters, err := charactersJSON(p.Characters)
	if err != nil {
		return principalRow{}, err
	}
	return principalRow{
		titleID:    titleID,
		ordering:   int64(p.Ordering),
		nameID:     nameID,
		category:   lookups.category.id(p.Category),
		job:        internedOrNull(lookups.job, p.Job),
		characters: characters,
	}, nil
}

// bindPrincipalRow appends a principals row's values in principalColumns order.
func bindPrincipalRow(args []any, r principalRow) []any {
	return append(args, r.titleID, r.ordering, r.nameID, r.category, r.job, r.characters)
}

// internedOrNull interns a non-empty value, mapping the reader's empty string
// (its \N) to nil so an absent value is stored as SQL NULL.
func internedOrNull(in *interner, value string) any {
	if value == "" {
		return nil
	}
	return in.id(value)
}

// charactersJSON re-encodes the reader's already-sanitised character names as a
// JSON array, or nil when the credit had none.
func charactersJSON(names []string) (any, error) {
	if names == nil {
		return nil, nil
	}
	data, err := json.Marshal(names)
	if err != nil {
		return nil, fmt.Errorf("encoding characters: %w", err)
	}
	return string(data), nil
}
