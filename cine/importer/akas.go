package importer

import (
	"context"
	"database/sql"
	"fmt"
	"io"

	"local.dev/cine/reader"
)

// akasColumns are the akas columns in the order bindAkasRow writes them.
var akasColumns = []string{
	"title_id", "ordering", "title", "region", "language", "types", "is_original_title",
}

// akaAttributeColumns are the akas_carry_attributes columns in bindAkaAttributeRow order.
var akaAttributeColumns = []string{"title_id", "ordering", "attribute_id"}

// akasLookups holds the interners populated while reading title.akas.
type akasLookups struct {
	region    *interner
	language  *interner
	akaType   *interner
	attribute *interner
}

// importAkas streams title.akas into the akas table, interning its region and
// language, folding its types into a bitmask, and fanning its attributes out
// into the akas_carry_attributes junction. Rows for titles the filter refuses are
// not written. It returns the number of akas rows written; the caller writes the
// returned lookups to their tables once the pass completes.
func importAkas(ctx context.Context, tx *sql.Tx, akas io.Reader, filter titleFilter) (counts, *akasLookups, error) {
	lookups := &akasLookups{
		region:    newInterner(),
		language:  newInterner(),
		akaType:   newInterner(),
		attribute: newInterner(),
	}
	titles, err := newBatchInserter(ctx, tx, "akas", akasColumns, bindAkasRow)
	if err != nil {
		return counts{}, nil, err
	}
	attributes, err := newBatchInserter(ctx, tx, "akas_carry_attributes", akaAttributeColumns, bindAkaAttributeRow)
	if err != nil {
		return counts{}, nil, err
	}
	var read int64
	for record, err := range reader.ReadTitleAkas(akas) {
		if err != nil {
			return counts{}, nil, err
		}
		read++
		titleID, err := parseID(record.TitleID)
		if err != nil {
			return counts{}, nil, rowError(read, record.TitleID, fmt.Errorf("titleId: %w", err))
		}
		// Refusing before the row is built keeps dropped values out of the interners.
		if !filter.allows(titleID) {
			continue
		}
		ordering := int64(record.Ordering)
		if err := titles.Add(ctx, buildAkasRow(record, titleID, ordering, lookups)); err != nil {
			return counts{}, nil, rowError(read, record.TitleID, err)
		}
		if err := addAttributes(ctx, attributes, titleID, ordering, record.Attributes, lookups.attribute); err != nil {
			return counts{}, nil, rowError(read, record.TitleID, err)
		}
	}
	if err := titles.Flush(ctx); err != nil {
		return counts{}, nil, err
	}
	if err := attributes.Flush(ctx); err != nil {
		return counts{}, nil, err
	}
	return counts{read: read, written: titles.Added()}, lookups, nil
}

// akasRow holds one akas row's values in column order; a nil field is stored as
// SQL NULL.
type akasRow struct {
	titleID         int64
	ordering        int64
	title           string
	region          any
	language        any
	types           int64
	isOriginalTitle int64
}

// buildAkasRow transforms a reader record into an akas row, interning its region
// and language and folding its types into a bitmask.
func buildAkasRow(a reader.TitleAkas, titleID, ordering int64, lookups *akasLookups) akasRow {
	var types int64
	for _, name := range a.Types {
		types |= lookups.akaType.bit(name)
	}
	return akasRow{
		titleID:         titleID,
		ordering:        ordering,
		title:           a.Title,
		region:          internedOrNull(lookups.region, a.Region),
		language:        internedOrNull(lookups.language, a.Language),
		types:           types,
		isOriginalTitle: boolToInt(a.IsOriginalTitle),
	}
}

// bindAkasRow appends an akas row's values in akasColumns order.
func bindAkasRow(args []any, r akasRow) []any {
	return append(args,
		r.titleID, r.ordering, r.title, r.region, r.language, r.types, r.isOriginalTitle)
}

// akaAttributeRow links one akas row to one of its attributes.
type akaAttributeRow struct {
	titleID     int64
	ordering    int64
	attributeID int64
}

// addAttributes adds one akas_carry_attributes row per attribute on the akas row.
func addAttributes(ctx context.Context, inserter *batchInserter[akaAttributeRow], titleID, ordering int64, attrs []string, attribute *interner) error {
	for _, name := range attrs {
		row := akaAttributeRow{titleID: titleID, ordering: ordering, attributeID: attribute.id(name)}
		if err := inserter.Add(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

// bindAkaAttributeRow appends an akas_carry_attributes row's values in column order.
func bindAkaAttributeRow(args []any, r akaAttributeRow) []any {
	return append(args, r.titleID, r.ordering, r.attributeID)
}
