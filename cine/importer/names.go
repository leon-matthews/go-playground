package importer

import (
	"context"
	"database/sql"
	"io"

	"local.dev/cine/reader"
)

// nameColumns are the names columns in the order bindNameRow writes them.
var nameColumns = []string{
	"id", "primary_name", "birth_year", "death_year", "primary_profession",
}

// importNames streams name.basics into the names table, interning the
// primary_profession bitmask. It returns the number of names written; the caller
// writes the returned interner to the profession lookup once the pass completes.
// knownForTitles is left for the opt-in name_known_for sub-layer.
func importNames(ctx context.Context, tx *sql.Tx, basics io.Reader) (int64, *interner, error) {
	profession := newInterner()
	inserter, err := newBatchInserter(ctx, tx, "names", nameColumns, bindNameRow)
	if err != nil {
		return 0, nil, err
	}
	for record, err := range reader.ReadNameBasics(basics) {
		if err != nil {
			return 0, nil, err
		}
		row, err := buildNameRow(record, profession)
		if err != nil {
			return 0, nil, err
		}
		if err := inserter.Add(ctx, row); err != nil {
			return 0, nil, err
		}
	}
	if err := inserter.Flush(ctx); err != nil {
		return 0, nil, err
	}
	return inserter.Added(), profession, nil
}

// nameRow holds one names row's values in column order; a nil field is stored as
// SQL NULL.
type nameRow struct {
	id                int64
	primaryName       string
	birthYear         any
	deathYear         any
	primaryProfession int64
}

// buildNameRow transforms a reader record into a names row, interning its
// professions into a bitmask.
func buildNameRow(n reader.NameBasics, profession *interner) (nameRow, error) {
	id, err := parseID(n.Nconst)
	if err != nil {
		return nameRow{}, err
	}
	var professions int64
	for _, name := range n.PrimaryProfession {
		professions |= profession.bit(name)
	}
	return nameRow{
		id:                id,
		primaryName:       n.PrimaryName,
		birthYear:         nullableInt(n.BirthYear),
		deathYear:         nullableInt(n.DeathYear),
		primaryProfession: professions,
	}, nil
}

// bindNameRow appends a names row's values in nameColumns order.
func bindNameRow(args []any, r nameRow) []any {
	return append(args, r.id, r.primaryName, r.birthYear, r.deathYear, r.primaryProfession)
}
