package pwned

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

const separator = ':'

// ETagStore is an in-memory store that can saved to a file
type ETagStore map[Prefix]string

func NewETagStore() ETagStore {
	return make(map[Prefix]string)
}

func (s ETagStore) Load(name string) error {
	fp, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("loading etags: %w", err)
	}

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
			return fmt.Errorf("loading etags, row %d: %w", i, err)
		}
		prefix, err := NewPrefix(row[0])
		if err != nil {
			return fmt.Errorf("loading etags, row %d: %w", i, err)
		}
		s[prefix] = row[1]
	}

	return nil
}

// Save writes map out to colon-separated file
func (s ETagStore) Save(name string) error {
	fp, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("saving etags: %w", err)
	}
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
