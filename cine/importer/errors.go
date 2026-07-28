package importer

import "fmt"

// Each level adds only the context the level below could not know, so a level
// with nothing of its own to add returns the error bare.

// rowError names the source row a pass failed on.
//
// The row is the pass's count of data rows, one less than the TSV line number.
// Reader errors are left bare: they carry that line number already.
func rowError(row int64, id string, err error) error {
	return fmt.Errorf("row %d (%s): %w", row, id, err)
}
