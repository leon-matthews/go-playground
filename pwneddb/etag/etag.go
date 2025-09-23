package etag

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

const separator = ':'

type ETagStore map[string]string

func NewETagStore() ETagStore {
	return make(map[string]string)
}

func (es ETagStore) Load(name string) error {
	fp, err := os.Open(name)
	if err != nil {
		return fmt.Errorf("loading etags: %w", err)
	}

	r := csv.NewReader(fp)
	r.Comma = separator
	i := 0
	for {
		row, err := r.Read()
		fmt.Printf("[%T]%+[1]v\n", row)
		i++
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("loading etags, row %d: %w", i, err)
		}
		es[row[0]] = row[1]
	}
	return nil
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
