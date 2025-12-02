package greetings

import (
	"regexp"
	"testing"
)

// Check for valid return values
func TestHelloName(t *testing.T) {
	name := "Leon"
	expected := regexp.MustCompile(`\b` + name + `\b`)

	// No match, or an error?
	msg, err := Hello("Leon")
	if !expected.MatchString(msg) || err != nil {
		t.Fatalf(`Hello("Gladys") = %q, %v, expected match for %#q, nil`, msg, err, expected)
	}
}

// Expected error if empty string passed to `Hello()`
func TestHelloEmpty(t *testing.T) {
	msg, err := Hello("")

	// Non-empty message, or missing error?
	if msg != "" || err == nil {
		t.Fatalf(`Hello("") = %q, %v, want "", error`, msg, err)
	}
}
