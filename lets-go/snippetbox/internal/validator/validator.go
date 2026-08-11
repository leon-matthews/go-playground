package validator

import (
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

// Email sanity check
// Pattern used as recommended by the W3C and WHATWG for validating email addresses:
// https://html.spec.whatwg.org/multipage/input.html#valid-e-mail-address
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// Validator holds map of error messages for each field
type Validator struct {
	FieldErrors    map[string]string
	NonFieldErrors []string
}

// Valid returns true if there are no errors
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0 && len(v.NonFieldErrors) == 0
}

// AddFieldError adds an error message for key, only if no existing error exists
func (v *Validator) AddFieldError(key, message string) {
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[key]; !exists {
		v.FieldErrors[key] = message
	}
}

// AddFieldError adds an error not tied to a form field
func (v *Validator) AddNonFieldError(message string) {
	v.NonFieldErrors = append(v.NonFieldErrors, message)
}

// CheckField adds error message only if validation check is not 'ok'
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

// NotBlank returns true if value is not an empty string
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// MinChars() returns true if a value contains n characters or more.
func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

// MaxChars() returns true if a value contains no more than n unicode characters.
func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// MaxBytes() returns true if a value contains n bytes or less.
func MaxBytes(value string, n int) bool {
	return len(value) <= n
}

// PermittedValue() returns true if a value is in given allow list
func PermittedValue[T comparable](value T, allowed ...T) bool {
	return slices.Contains(allowed, value)
}

// Matches() returns true if a value matches a provided compiled regular expression
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
