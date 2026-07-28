package importer

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	"local.dev/cine/reader"
)

// Crew roles stored in the crew.role column.
const (
	roleDirector = 0
	roleWriter   = 1
)

// crewColumns are the crew columns in the order bindCrewRow writes them.
var crewColumns = []string{"title_id", "name_id", "role", "position"}

// crewRow is one director or writer credit.
type crewRow struct {
	titleID  int64
	nameID   int64
	role     int64
	position int64
}

// importCrew streams title.crew into the titles_credit_names table, fanning each
// row's director and writer lists out into one role-tagged row per person. Credits
// for titles the filter refuses are not written. It returns the number of credit
// rows written.
func importCrew(ctx context.Context, tx *sql.Tx, crew io.Reader, filter titleFilter) (counts, error) {
	inserter, err := newBatchInserter(ctx, tx, "titles_credit_names", crewColumns, bindCrewRow)
	if err != nil {
		return counts{}, err
	}
	var read int64
	for record, err := range reader.ReadTitleCrew(crew) {
		if err != nil {
			return counts{}, err
		}
		read++
		titleID, err := parseID(record.Tconst)
		if err != nil {
			return counts{}, rowError(read, record.Tconst, fmt.Errorf("tconst: %w", err))
		}
		if !filter.allows(titleID) {
			continue
		}
		if err := addCrew(ctx, inserter, titleID, record.Directors, roleDirector, "directors"); err != nil {
			return counts{}, rowError(read, record.Tconst, err)
		}
		if err := addCrew(ctx, inserter, titleID, record.Writers, roleWriter, "writers"); err != nil {
			return counts{}, rowError(read, record.Tconst, err)
		}
	}
	if err := inserter.Flush(ctx); err != nil {
		return counts{}, err
	}
	return counts{read: read, written: inserter.Added()}, nil
}

// addCrew adds one crew row per person in the list, all tagged with role and
// numbered from one so that IMDb's order within the list survives. column is the
// title.crew column role came from, for its errors.
func addCrew(ctx context.Context, inserter *batchInserter[crewRow], titleID int64, people []string, role int64, column string) error {
	for i, nconst := range people {
		nameID, err := parseID(nconst)
		if err != nil {
			return fmt.Errorf("%s[%d]: %w", column, i+1, err)
		}
		row := crewRow{titleID: titleID, nameID: nameID, role: role, position: int64(i + 1)}
		if err := inserter.Add(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

// bindCrewRow appends a crew row's values in crewColumns order.
func bindCrewRow(args []any, r crewRow) []any {
	return append(args, r.titleID, r.nameID, r.role, r.position)
}
