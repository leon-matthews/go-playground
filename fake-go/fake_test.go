package fake

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestReproducible checks that a given seed always yields the same sequence.
func TestReproducible(t *testing.T) {
	sample := func(seed uint64) string {
		f := New(seed)
		var b strings.Builder
		fmt.Fprintln(&b, f.Int(1, 100))
		fmt.Fprintln(&b, f.Price(1, 100))
		fmt.Fprintln(&b, f.FullName())
		fmt.Fprintln(&b, f.Words(3, 8))
		fmt.Fprintln(&b, f.Address())
		return b.String()
	}
	first, second := sample(7), sample(7)
	if first != second {
		t.Error("same seed produced different output")
	}
	if first == sample(8) {
		t.Error("different seeds produced identical output")
	}
}

func TestInt(t *testing.T) {
	f := New(1)
	for range 1000 {
		if n := f.Int(5, 10); n < 5 || n > 10 {
			t.Fatalf("Int out of range: %d", n)
		}
	}
}

func TestIntPanicsOnBadRange(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected panic for high <= low")
		}
	}()
	New(1).Int(10, 5)
}

func TestFloat(t *testing.T) {
	f := New(1)
	for range 1000 {
		if v := f.Float(1.5, 2.5); v < 1.5 || v >= 2.5 {
			t.Fatalf("Float out of range: %v", v)
		}
	}
}

func TestPrice(t *testing.T) {
	f := New(1)
	for range 1000 {
		if cents := f.Price(1, 100); cents < 100 || cents > 10000 {
			t.Fatalf("Price out of range: %d", cents)
		}
	}
}

func TestDigits(t *testing.T) {
	f := New(1)
	for range 1000 {
		s := f.Digits(4, 4)
		if len(s) != 4 {
			t.Fatalf("Digits wrong length: %q", s)
		}
		if s == "0000" {
			t.Fatal("Digits returned all zeros")
		}
		for _, r := range s {
			if r < '0' || r > '9' {
				t.Fatalf("Digits contains non-digit: %q", s)
			}
		}
	}
}

func TestLetters(t *testing.T) {
	f := New(1)
	s := f.Letters(8, 8)
	if len(s) != 8 {
		t.Fatalf("Letters wrong length: %q", s)
	}
	for _, r := range s {
		if r < 'A' || r > 'Z' {
			t.Fatalf("Letters contains non-upper-case: %q", s)
		}
	}
}

func TestCode(t *testing.T) {
	f := New(1)
	code := f.Code(6, 6)
	if len(code) != 6 {
		t.Fatalf("Code wrong length: %q", code)
	}
}

func TestWordsCount(t *testing.T) {
	f := New(1)
	if got := len(strings.Fields(f.Words(5, 5))); got != 5 {
		t.Fatalf("Words: got %d words, want 5", got)
	}
}

func TestParagraphsCount(t *testing.T) {
	f := New(1)
	got := strings.Count(f.Paragraphs(3, 3), "\n\n") + 1
	if got != 3 {
		t.Fatalf("Paragraphs: got %d paragraphs, want 3", got)
	}
}

func TestParagraphsHTML(t *testing.T) {
	f := New(1)
	html := f.ParagraphsHTML(2, 2)
	if strings.Count(html, "<p>") != 2 || strings.Count(html, "</p>") != 2 {
		t.Fatalf("ParagraphsHTML missing tags: %q", html)
	}
}

func TestPhone(t *testing.T) {
	pattern := regexp.MustCompile(`^0\d{1,2}[- ]?555[- ]?\d{4}$`)
	f := New(1)
	for range 100 {
		if phone := f.Phone(); !pattern.MatchString(phone) {
			t.Fatalf("Phone bad format: %q", phone)
		}
	}
}

func TestEmailFor(t *testing.T) {
	if got := New(1).EmailFor("John Smith"); got != "john.smith@example.com" {
		t.Fatalf("EmailFor: got %q", got)
	}
}

func TestWebsiteFor(t *testing.T) {
	if got := New(1).WebsiteFor("Acme Widgets"); got != "https://acme.widgets.com/" {
		t.Fatalf("WebsiteFor: got %q", got)
	}
}

func TestAddress(t *testing.T) {
	f := New(1)
	for range 100 {
		a := f.Address()
		if a.Address1 == "" || a.City == "" {
			t.Fatal("Address missing required field")
		}
		if len(a.PostCode) != 4 {
			t.Fatalf("Address postcode not 4 digits: %q", a.PostCode)
		}
	}
}

func TestStreetUnitLetter(t *testing.T) {
	pattern := regexp.MustCompile(`^\d+([A-Z]?) `)
	f := New(1)
	sawA := false
	for range 2000 {
		street := f.Street()
		m := pattern.FindStringSubmatch(street)
		if m == nil {
			t.Fatalf("Street bad format: %q", street)
		}
		switch unit := m[1]; {
		case unit == "":
			continue
		case unit < "A" || unit > "E":
			t.Fatalf("Street unit letter out of range A-E: %q", unit)
		case unit == "A":
			sawA = true
		}
	}
	if !sawA {
		t.Error("Street never produced unit letter A (off-by-one regression)")
	}
}

func TestDateOfBirth(t *testing.T) {
	f := New(1)
	now := time.Now()
	for range 100 {
		years := now.Sub(f.DateOfBirth(18, 65)).Seconds() / secondsPerYear
		if years < 18-0.1 || years > 65+0.1 {
			t.Fatalf("DateOfBirth age out of range: %.2f", years)
		}
	}
}

func TestRelativeTimeNow(t *testing.T) {
	got, err := New(1).RelativeTime("now")
	if err != nil {
		t.Fatal(err)
	}
	if d := time.Since(got); d < 0 || d > time.Second {
		t.Fatalf("RelativeTime(now) off by %v", d)
	}
}

func TestRelativeTimeRelative(t *testing.T) {
	got, err := New(1).RelativeTime("-1 year")
	if err != nil {
		t.Fatal(err)
	}
	years := time.Since(got).Seconds() / secondsPerYear
	if years < 0.99 || years > 1.01 {
		t.Fatalf("RelativeTime(-1 year) gave %.3f years ago", years)
	}
}

func TestRelativeTimeError(t *testing.T) {
	if _, err := New(1).RelativeTime("3 fortnights"); err == nil {
		t.Error("expected error for unknown unit")
	}
}

func TestBetween(t *testing.T) {
	f := New(1)
	start := time.Now()
	end := start.Add(48 * time.Hour)
	for range 100 {
		got := f.Between(start, end)
		if got.Before(start) || !got.Before(end) {
			t.Fatalf("Between out of range: %v", got)
		}
	}
}

func TestParsePairs(t *testing.T) {
	got, err := parsePairs("2y4w7d")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Duration((2*secondsPerYear + 4*secondsPerWeek + 7*secondsPerDay) * float64(time.Second))
	if got != want {
		t.Fatalf("parsePairs: got %v, want %v", got, want)
	}
}

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"John Smith":      "john-smith",
		"  Hello!!World ": "hello-world",
		"Once & Again":    "once-again",
		"Café René":       "cafe-rene",
		"Ōtautahi":        "otautahi",
	}
	for in, want := range cases {
		if got := slug(in, 50); got != want {
			t.Errorf("slug(%q): got %q, want %q", in, got, want)
		}
	}
}
