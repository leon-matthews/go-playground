package imdb

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// missing is IMDb's marker for a missing field.
const missing = `\N`

// optString maps IMDb's \N to the empty string and passes anything else through.
func optString(s string) string {
	if s == missing {
		return ""
	}
	return s
}

// optInt parses an optional integer field, mapping \N to nil.
func optInt(s string) (*int, error) {
	if s == missing {
		return nil, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("invalid integer %q: %w", s, err)
	}
	return &n, nil
}

// reqInt parses a required integer field.
func reqInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid integer %q: %w", s, err)
	}
	return n, nil
}

// parseFloat parses a required floating-point field.
func parseFloat(s string) (float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q: %w", s, err)
	}
	return f, nil
}

// parseBool parses IMDb's "0" and "1" boolean fields.
func parseBool(s string) (bool, error) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("invalid boolean %q", s)
	}
	return b, nil
}

// splitList splits a comma-separated IMDb list; \N or empty yields nil.
func splitList(s string) []string {
	if s == missing || s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

// parseCharacters decodes the JSON string array held in the principals
// characters field; \N yields nil.
func parseCharacters(s string) ([]string, error) {
	if s == missing {
		return nil, nil
	}
	var names []string
	if err := json.Unmarshal([]byte(s), &names); err != nil {
		return nil, fmt.Errorf("invalid characters %q: %w", s, err)
	}
	return names, nil
}

// cursor extracts typed values from one row's tab-separated fields, keeping the
// first parse error so a record can be built in a single declarative block.
// The read core guarantees the field count, so indexing is always in range.
type cursor struct {
	fields []string
	err    error
}

func (c *cursor) str(i int) string    { return c.fields[i] }
func (c *cursor) optStr(i int) string { return optString(c.fields[i]) }
func (c *cursor) list(i int) []string { return splitList(c.fields[i]) }

func (c *cursor) optInt(i int) *int {
	if c.err != nil {
		return nil
	}
	n, err := optInt(c.fields[i])
	c.keep(err)
	return n
}

func (c *cursor) reqInt(i int) int {
	if c.err != nil {
		return 0
	}
	n, err := reqInt(c.fields[i])
	c.keep(err)
	return n
}

func (c *cursor) boolean(i int) bool {
	if c.err != nil {
		return false
	}
	b, err := parseBool(c.fields[i])
	c.keep(err)
	return b
}

func (c *cursor) float(i int) float64 {
	if c.err != nil {
		return 0
	}
	f, err := parseFloat(c.fields[i])
	c.keep(err)
	return f
}

func (c *cursor) characters(i int) []string {
	if c.err != nil {
		return nil
	}
	names, err := parseCharacters(c.fields[i])
	c.keep(err)
	return names
}

// keep records the first error seen.
func (c *cursor) keep(err error) {
	if c.err == nil {
		c.err = err
	}
}
