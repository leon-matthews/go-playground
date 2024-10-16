package concatenation

import (
	"testing"
)

func makeSlice() []string {
	return []string{"a", "bb", "ccc", "dddd", "eeeee", "ffffff"}
}

func TestConcatenation(t *testing.T) {
	parts := makeSlice()
	want := "a bb ccc dddd eeeee ffffff"

	t.Run("ConcatOperator", func(t *testing.T) {
		got := ConcatOperator(parts...)
		assertStrings(t, got, want)
	})

	t.Run("ConcatBuilder", func(t *testing.T) {
		got := ConcatBuilder(parts...)
		assertStrings(t, got, want)
	})

	t.Run("ConcatJoin", func(t *testing.T) {
		got := ConcatJoin(parts...)
		assertStrings(t, got, want)
	})
}

func BenchmarkConcatOperator(b *testing.B) {
	parts := makeSlice()
	for i := 0; i < b.N; i++ {
		ConcatOperator(parts...)
	}
}

func BenchmarkConcatBuilder(b *testing.B) {
	parts := makeSlice()
	for i := 0; i < b.N; i++ {
		ConcatBuilder(parts...)
	}
}

func BenchmarkConcatJoin(b *testing.B) {
	parts := makeSlice()
	for i := 0; i < b.N; i++ {
		ConcatJoin(parts...)
	}
}

func assertStrings(t testing.TB, got string, want string) {
	t.Helper()
	if got != want {
		t.Errorf("expected %q but got %q", want, got)
	}
}
