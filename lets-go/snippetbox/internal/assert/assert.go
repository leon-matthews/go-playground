// Package assert provides assertion helpers for tests.
//
// Each helper reports a test error via t.Errorf, so execution continues after a
// failed assertion.
package assert

import (
	"reflect"
	"testing"
)

// Equal reports a test error unless got and want are deeply equal.
func Equal[A any](t *testing.T, got, want A) {
	t.Helper()
	if !isEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

// NotEqual reports a test error if got and want are equal.
func NotEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()

	if got == want {
		t.Errorf("got: %v; expected values to be different", got)
	}
}

// True reports a test error unless got is true.
func True(t *testing.T, got bool) {
	t.Helper()
	if !got {
		t.Errorf("got: %v, want: true", got)
	}
}

// False reports a test error unless got is false.
func False(t *testing.T, got bool) {
	t.Helper()

	if got {
		t.Errorf("got: true; want: false")
	}
}

// Nil reports a test error unless got is nil, including an interface holding a nil pointer.
func Nil(t *testing.T, got any) {
	t.Helper()
	if !isNil(got) {
		t.Errorf("got: %v; want: nil", got)
	}
}

// NotNil reports a test error if got is nil, including an interface holding a nil pointer.
func NotNil(t *testing.T, got any) {
	t.Helper()

	if isNil(got) {
		t.Errorf("got: %v; want: non-nil", got)
	}
}

// isEqual reports whether got and want are deeply equal, treating two nils as equal.
func isEqual[A any](got, want A) bool {
	if isNil(got) && isNil(want) {
		return true
	}
	return reflect.DeepEqual(got, want)
}

// isNil reports whether v is nil, or holds a nil channel, func, map, pointer, or slice.
func isNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	}

	// Other types like string, bool, int are never nil.
	return false
}
