package pwned

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

const separator = ':'

// ETags is an in-memory store that can saved to a file
type ETags map[Prefix]string

// NewETags builds an empty ETags mapping prefixes to ETag
func NewETags() ETags {
	return make(map[Prefix]string)
}

// ETagsLoad attempts to load ETags from colon-separated file
func ETagsLoad(name string) (ETags, error) {
	fp, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("loading etags: %w", err)
	}

	const expectedRows = 2
	etags := make(ETags)
	r := csv.NewReader(fp)
	r.Comma = separator
	i := 0
	for {
		row, err := r.Read()
		i++
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("loading etags, row %d: %w", i, err)
		}

		if len(row) != expectedRows {
			return nil, fmt.Errorf("loading etags, wrong number of rows on row %d: %w", i, err)
		}

		prefix, err := NewPrefix(row[0])
		if err != nil {
			return nil, fmt.Errorf("loading etags, bad prefix on row %d: %w", i, err)
		}
		etags[prefix] = row[1]
	}

	return etags, nil
}

// Save writes map out to colon-separated file
func (s ETags) Save(path string) error {
	fp, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("saving etags: %w", err)
	}
	defer fp.Close()

	w := csv.NewWriter(fp)
	w.Comma = separator
	for k, v := range s {
		if err := w.Write([]string{string(k), v}); err != nil {
			return fmt.Errorf("writing etag field (%v:%v): %w", k, v, err)
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("saving etags: %w", err)
	}
	return nil
}
