package loremipsum

import (
	"math/rand/v2"
	"strings"
	"testing"
)

// newRNG returns a deterministically seeded generator for tests.
func newRNG() *rand.Rand {
	return rand.New(rand.NewPCG(1, 1))
}

func TestWordsCount(t *testing.T) {
	rng := newRNG()
	for _, count := range []int{1, 5, 19, 20, 200, 500} {
		if got := len(strings.Fields(Words(rng, count, false))); got != count {
			t.Errorf("Words(%d): got %d words", count, got)
		}
	}
}

func TestWordsCommon(t *testing.T) {
	got := Words(newRNG(), 5, true)
	if want := "lorem ipsum dolor sit amet"; got != want {
		t.Errorf("Words common: got %q, want %q", got, want)
	}
}

func TestSampleUnique(t *testing.T) {
	got := sample(newRNG(), words, 12)
	if len(got) != 12 {
		t.Fatalf("sample returned %d words, want 12", len(got))
	}
	seen := make(map[string]bool, len(got))
	for _, word := range got {
		if seen[word] {
			t.Errorf("sample returned duplicate word: %q", word)
		}
		seen[word] = true
	}
}

func TestParagraphsCommon(t *testing.T) {
	got := Paragraphs(newRNG(), 3, true)
	if len(got) != 3 {
		t.Fatalf("Paragraphs returned %d, want 3", len(got))
	}
	if got[0] != CommonP {
		t.Error("first paragraph is not the common paragraph")
	}
}

func TestSentenceShape(t *testing.T) {
	s := Sentence(newRNG())
	if last := s[len(s)-1]; last != '.' && last != '?' {
		t.Errorf("sentence does not end in . or ?: %q", s)
	}
	if first := s[0]; first < 'A' || first > 'Z' {
		t.Errorf("sentence not capitalised: %q", s)
	}
}
