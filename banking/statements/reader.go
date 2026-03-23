// Package statements provides a common interface for reading bank statements.
package statements

import (
	"fmt"
	"strings"

	"banking/common"
)

// Format describes a bank statement format that can be detected and read.
type Format interface {
	Name() string
	Detect(data []byte) error
	Read(data []byte) ([]*common.Transaction, error)
}

var formats []Format

// Register adds a format to the registry.
func Register(f Format) {
	formats = append(formats, f)
}

// Get looks up a format by name.
func Get(name string) (Format, bool) {
	for _, f := range formats {
		if f.Name() == name {
			return f, true
		}
	}
	return nil, false
}

// Detect returns the first format that recognises the given data.
func Detect(data []byte) (Format, error) {
	var reasons []string
	for _, ff := range formats {
		if err := ff.Detect(data); err == nil {
			return ff, nil
		} else {
			reasons = append(reasons, fmt.Sprintf("  %s: %s", ff.Name(), err))
		}
	}
	return nil, fmt.Errorf("unrecognised statement format\n%s", strings.Join(reasons, "\n"))
}

// Names returns the names of all registered formats.
func Names() []string {
	names := make([]string, len(formats))
	for i, f := range formats {
		names[i] = f.Name()
	}
	return names
}
