package importer

import (
	"context"
	"database/sql"
	"io"

	"local.dev/cine/reader"
)

// Crew roles stored in the crew.role column.
const (
	roleDirector = 0
	roleWriter   = 1
)

// crewColumns are the crew columns in the order bindCrewRow writes them.
var crewColumns = []string{"title_id", "name_id", "role"}

// crewRow is one director or writer credit.
type crewRow struct {
	titleID int64
	nameID  int64
	role    int64
}

// importCrew streams title.crew into the crew table, fanning each row's director
// and writer lists out into one role-tagged row per person.
func importCrew(ctx context.Context, tx *sql.Tx, crew io.Reader) error {
	inserter, err := newBatchInserter(ctx, tx, "crew", crewColumns, bindCrewRow)
	if err != nil {
		return err
	}
	for record, err := range reader.ReadTitleCrew(crew) {
		if err != nil {
			return err
		}
		titleID, err := parseID(record.Tconst)
		if err != nil {
			return err
		}
		if err := addCrew(ctx, inserter, titleID, record.Directors, roleDirector); err != nil {
			return err
		}
		if err := addCrew(ctx, inserter, titleID, record.Writers, roleWriter); err != nil {
			return err
		}
	}
	return inserter.Flush(ctx)
}

// addCrew adds one crew row per person in the list, all tagged with role.
func addCrew(ctx context.Context, inserter *batchInserter[crewRow], titleID int64, people []string, role int64) error {
	for _, nconst := range people {
		nameID, err := parseID(nconst)
		if err != nil {
			return err
		}
		if err := inserter.Add(ctx, crewRow{titleID: titleID, nameID: nameID, role: role}); err != nil {
			return err
		}
	}
	return nil
}

// bindCrewRow appends a crew row's values in crewColumns order.
func bindCrewRow(args []any, r crewRow) []any {
	return append(args, r.titleID, r.nameID, r.role)
}
