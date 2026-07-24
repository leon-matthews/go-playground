package reader

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// nullMarker is the token IMDb uses in the TSV files for an absent field.
const nullMarker = `\N`

// missing is the value of an absent optional integer field, standing in for
// IMDb's \N. Code writing these records to a database must map missing to SQL
// NULL rather than storing it as -1.
const missing = -1

// optionalString maps IMDb's \N to the empty string and passes anything else through.
func optionalString(s string) string {
	if s == nullMarker {
		return ""
	}
	return s
}

// optionalInt parses an optional integer field, mapping \N to missing.
func optionalInt(s string) (int, error) {
	if s == nullMarker {
		return missing, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return missing, fmt.Errorf("invalid integer %q: %w", s, err)
	}
	return n, nil
}

// requiredInt parses a required integer field.
func requiredInt(s string) (int, error) {
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
	if s == nullMarker || s == "" {
		return nil
	}
	return strings.Split(s, ",")
}

// parseCharacters decodes the JSON string array in the principals characters field.
// A literal "\N" input returns nil.
func parseCharacters(s string) ([]string, error) {
	if s == nullMarker {
		return nil, nil
	}

	// Try to use fast-path parser
	if names, ok := fastUnmarshal(s); ok {
		return names, nil
	}
	var names []string
	if err := json.Unmarshal([]byte(s), &names); err != nil {
		return nil, fmt.Errorf("invalid characters %q: %w", s, err)
	}
	return names, nil
}

// fastUnmarshal parses a plain ["a","b"] JSON array into a slice of strings.
// It returns ok false for anything challenging. The caller should then defer
// to encoding/json for more robust handling.
func fastUnmarshal(s string) (names []string, ok bool) {
	if len(s) < 4 || !strings.HasPrefix(s, `["`) || !strings.HasSuffix(s, `"]`) {
		return nil, false
	}
	// Escapes and invalid UTF-8 both change what encoding/json would produce
	if strings.IndexByte(s, '\\') >= 0 || !utf8.ValidString(s) {
		return nil, false
	}
	// No backslash means every quote is a delimiter, so "," separates elements
	names = strings.Split(s[2:len(s)-2], `","`)
	for _, name := range names {
		// A bare quote or control byte is invalid JSON inside a string, so
		// defer those to the fallback and keep encoding/json authoritative
		for i := 0; i < len(name); i++ {
			if b := name[i]; b == '"' || b < 0x20 {
				return nil, false
			}
		}
	}
	return names, true
}
