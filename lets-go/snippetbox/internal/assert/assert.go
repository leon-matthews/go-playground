package assert

import (
	"reflect"
	"testing"
)

func Equal[A any](t *testing.T, got, want A) {
	t.Helper()
	if !isEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

func True(t *testing.T, got bool) {
	t.Helper()
	if !got {
		t.Errorf("got: %v, want: true", got)
	}
}

func Nil(t *testing.T, got any) {
	t.Helper()
	if !isNil(got) {
		t.Errorf("got: %v; want: nil", got)
	}
}

func isEqual[A any](got, want A) bool {
	if isNil(got) && isNil(want) {
		return true
	}
	return reflect.DeepEqual(got, want)
}

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
