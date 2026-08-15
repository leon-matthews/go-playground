// Package assert contains helpers for unit tests
package assert

import (
	"reflect"
	"testing"
)

// Equal asserts that the generic types are equal
func Equal[A any](t *testing.T, got, want A) {
	t.Helper()
	if !isEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

// True asserts the bool is true
func True(t *testing.T, got bool) {
	t.Helper()
	if !got {
		t.Errorf("got: %v, want: true", got)
	}
}

// Nil asserts that got is nil, or a nil interface
func Nil(t *testing.T, got any) {
	t.Helper()
	if !isNil(got) {
		t.Errorf("got: %v; want: nil", got)
	}
}

// isEqual uses reflect.DeepEqual to implement generic equality test
func isEqual[A any](got, want A) bool {
	if isNil(got) && isNil(want) {
		return true
	}
	return reflect.DeepEqual(got, want)
}

// isNil also returns true if an interface's underlying type is nil
func isNil(v any) bool {
	if v == nil {
		return true
	}

	// Use reflection to check the underlying type for a nullable type
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	}

	// Other types like string, bool, int are never nil.
	return false
}
