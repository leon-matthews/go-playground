package etag

import (
	"encoding/csv"
	"fmt"
	"os"
)

const separator = ':'

type ETagStore map[string]string

func NewETagStore() ETagStore {
	return make(map[string]string)
}

// Save writes map out to colon-separated file
func (es ETagStore) Save(name string) error {
	fp, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("saving etags: %w", err)
	}
	w := csv.NewWriter(fp)
	w.Comma = separator
	for k, v := range es {
		if err := w.Write([]string{k, v}); err != nil {
			return fmt.Errorf("writing etag field (%v:%v): %w", k, v, err)
		}
	}
	w.Flush()

	if err := w.Error(); err != nil {
		return fmt.Errorf("saving etags: %w", err)
	}
	return nil
}
