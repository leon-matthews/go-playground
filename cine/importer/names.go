package importer

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	"local.dev/cine/reader"
)

// nameColumns are the names columns in the order bindNameRow writes them.
var nameColumns = []string{"id", "primary_name", "birth_year", "death_year"}

// importNames streams name.basics into the names table, fanning its two ordered
// list fields out into the names_primary_professions and names_known_for_titles
// junctions. It returns the number of names written (not junction rows); the
// caller writes the returned interner to the profession lookup once the pass
// completes. Every person is kept; only the known-for junction is title-keyed,
// so the filter reaches that alone.
func importNames(ctx context.Context, tx *sql.Tx, basics io.Reader, filter titleFilter) (counts, *interner, error) {
	profession := newInterner()
	inserter, err := newBatchInserter(ctx, tx, "names", nameColumns, bindNameRow)
	if err != nil {
		return counts{}, nil, err
	}
	professions, err := newBatchInserter(ctx, tx, "names_primary_professions", nameProfessionColumns, bindNameProfessionRow)
	if err != nil {
		return counts{}, nil, err
	}
	knownFor, err := newBatchInserter(ctx, tx, "names_known_for_titles", nameKnownForColumns, bindNameKnownForRow)
	if err != nil {
		return counts{}, nil, err
	}
	var read int64
	for record, err := range reader.ReadNameBasics(basics) {
		if err != nil {
			return counts{}, nil, err
		}
		read++
		row, err := buildNameRow(record)
		if err != nil {
			return counts{}, nil, rowError(read, record.Nconst, err)
		}
		if err := inserter.Add(ctx, row); err != nil {
			return counts{}, nil, rowError(read, record.Nconst, err)
		}
		if err := addProfessions(ctx, professions, row.id, record.PrimaryProfession, profession); err != nil {
			return counts{}, nil, rowError(read, record.Nconst, err)
		}
		if err := addKnownFor(ctx, knownFor, row.id, record.KnownForTitles, filter); err != nil {
			return counts{}, nil, rowError(read, record.Nconst, err)
		}
	}
	if err := inserter.Flush(ctx); err != nil {
		return counts{}, nil, err
	}
	if err := professions.Flush(ctx); err != nil {
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
	id          int64
	primaryName any
	birthYear   any
	deathYear   any
}

// buildNameRow transforms a reader record into a names row.
func buildNameRow(n reader.NameBasics) (nameRow, error) {
	id, err := parseID(n.Nconst)
	if err != nil {
		return nameRow{}, fmt.Errorf("nconst: %w", err)
	}
	return nameRow{
		id:          id,
		primaryName: nullableStr(n.PrimaryName),
		birthYear:   nullableInt(n.BirthYear),
		deathYear:   nullableInt(n.DeathYear),
	}, nil
}

// bindNameRow appends a names row's values in nameColumns order.
func bindNameRow(args []any, r nameRow) []any {
	return append(args, r.id, r.primaryName, r.birthYear, r.deathYear)
}

// nameProfessionColumns are the names_primary_professions columns in the order
// bindNameProfessionRow writes them.
var nameProfessionColumns = []string{"name_id", "position", "profession_id"}

// nameProfessionRow is one entry of a person's profession list.
type nameProfessionRow struct {
	nameID       int64
	position     int64
	professionID int64
}

// addProfessions adds one junction row per profession, numbering them from one so
// that IMDb's ranking of the list survives, and interning each name.
func addProfessions(ctx context.Context, inserter *batchInserter[nameProfessionRow], nameID int64, names []string, profession *interner) error {
	for i, name := range names {
		row := nameProfessionRow{
			nameID:       nameID,
			position:     int64(i + 1),
			professionID: profession.id(name),
		}
		if err := inserter.Add(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

// bindNameProfessionRow appends a names_primary_professions row's values in
// nameProfessionColumns order.
func bindNameProfessionRow(args []any, r nameProfessionRow) []any {
	return append(args, r.nameID, r.position, r.professionID)
}

// nameKnownForColumns are the names_known_for_titles columns in the order
// bindNameKnownForRow writes them.
var nameKnownForColumns = []string{"name_id", "position", "title_id"}

// nameKnownForRow is one entry of a person's known-for list.
type nameKnownForRow struct {
	nameID   int64
	position int64
	titleID  int64
}

// addKnownFor adds one junction row per known-for title the filter allows.
//
// Numbering is from one so IMDb's order survives; a refused title leaves a gap.
func addKnownFor(ctx context.Context, inserter *batchInserter[nameKnownForRow], nameID int64, titles []string, filter titleFilter) error {
	for i, tconst := range titles {
		titleID, err := parseID(tconst)
		if err != nil {
			return fmt.Errorf("knownForTitles[%d]: %w", i+1, err)
		}
		if !filter.allows(titleID) {
			continue
		}
		row := nameKnownForRow{nameID: nameID, position: int64(i + 1), titleID: titleID}
		if err := inserter.Add(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

// bindNameKnownForRow appends a names_known_for_titles row's values in
// nameKnownForColumns order.
func bindNameKnownForRow(args []any, r nameKnownForRow) []any {
	return append(args, r.nameID, r.position, r.titleID)
}
