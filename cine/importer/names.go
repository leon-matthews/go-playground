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
// primary_profession bitmask and fanning knownForTitles out into the
// name_known_for junction. It returns the number of names written (not junction
// rows); the caller writes the returned interner to the profession lookup once
// the pass completes.
func importNames(ctx context.Context, tx *sql.Tx, basics io.Reader) (counts, *interner, error) {
	profession := newInterner()
	inserter, err := newBatchInserter(ctx, tx, "names", nameColumns, bindNameRow)
	if err != nil {
		return counts{}, nil, err
	}
	knownFor, err := newBatchInserter(ctx, tx, "name_known_for", nameKnownForColumns, bindNameKnownForRow)
	if err != nil {
		return counts{}, nil, err
	}
	var read int64
	for record, err := range reader.ReadNameBasics(basics) {
		if err != nil {
			return counts{}, nil, err
		}
		read++
		row, err := buildNameRow(record, profession)
		if err != nil {
			return counts{}, nil, err
		}
		if err := inserter.Add(ctx, row); err != nil {
			return counts{}, nil, err
		}
		if err := addKnownFor(ctx, knownFor, row.id, record.KnownForTitles); err != nil {
			return counts{}, nil, err
		}
	}
	if err := inserter.Flush(ctx); err != nil {
		return counts{}, nil, err
	}
	if err := knownFor.Flush(ctx); err != nil {
		return counts{}, nil, err
	}
	return counts{read: read, written: inserter.Added()}, profession, nil
}

// nameRow holds one names row's values in column order; a nil field is stored as
// SQL NULL.
type nameRow struct {
	id                int64
	primaryName       any
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
		primaryName:       nullableStr(n.PrimaryName),
		birthYear:         nullableInt(n.BirthYear),
		deathYear:         nullableInt(n.DeathYear),
		primaryProfession: professions,
	}, nil
}

// bindNameRow appends a names row's values in nameColumns order.
func bindNameRow(args []any, r nameRow) []any {
	return append(args, r.id, r.primaryName, r.birthYear, r.deathYear, r.primaryProfession)
}

// nameKnownForColumns are the name_known_for columns in the order
// bindNameKnownForRow writes them.
var nameKnownForColumns = []string{"name_id", "position", "title_id"}

// nameKnownForRow is one entry of a person's known-for list.
type nameKnownForRow struct {
	nameID   int64
	position int64
	titleID  int64
}

// addKnownFor adds one junction row per known-for title, numbering them from one
// so that IMDb's ordering of the list survives.
func addKnownFor(ctx context.Context, inserter *batchInserter[nameKnownForRow], nameID int64, titles []string) error {
	for i, tconst := range titles {
		titleID, err := parseID(tconst)
		if err != nil {
			return err
		}
		row := nameKnownForRow{nameID: nameID, position: int64(i + 1), titleID: titleID}
		if err := inserter.Add(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

// bindNameKnownForRow appends a name_known_for row's values in
// nameKnownForColumns order.
func bindNameKnownForRow(args []any, r nameKnownForRow) []any {
	return append(args, r.nameID, r.position, r.titleID)
}
