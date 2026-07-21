package fake

import (
	"regexp"
	"strings"

	"local.dev/fake-go/loremipsum"
)

// paragraphBreak splits text into paragraphs on runs of two or more newlines.
var paragraphBreak = regexp.MustCompile(`\n{2,}`)

// Code returns an upper-case alphanumeric code, for example "DC4563".
//
// Its length is a fixed count when low equals high, otherwise a random length
// in the inclusive range [low, high].
func (f *Faker) Code(low, high int) string {
	length := f.count(low, high)
	numLetters := f.count(0, length/2)
	return f.letters(numLetters) + f.digits(length-numLetters)
}

// Letters returns a random upper-case string.
//
// Its length is a fixed count when low equals high, otherwise a random length
// in the inclusive range [low, high].
func (f *Faker) Letters(low, high int) string {
	return f.letters(f.count(low, high))
}

// Digits returns a random string of decimal digits, for example "0530".
//
// Its length is a fixed count when low equals high, otherwise a random length
// in the inclusive range [low, high]. The value is never all zeros. It panics
// if the length is not in the range 1 to 256.
func (f *Faker) Digits(low, high int) string {
	return f.digits(f.count(low, high))
}

// Word returns a single random lorem ipsum word.
func (f *Faker) Word() string {
	return loremipsum.Words(f.rng, 1, false)
}

// Words returns lorem ipsum words separated by single spaces.
//
// The number of words is a fixed count when low equals high, otherwise a random
// count in the inclusive range [low, high].
func (f *Faker) Words(low, high int) string {
	return loremipsum.Words(f.rng, f.count(low, high), false)
}

// Paragraph returns a single plain-text paragraph without line breaks.
func (f *Faker) Paragraph() string {
	return loremipsum.Paragraphs(f.rng, 1, false)[0]
}

// Paragraphs returns plain-text paragraphs separated by a blank line.
//
// The number of paragraphs is a fixed count when low equals high, otherwise a
// random count in the inclusive range [low, high].
func (f *Faker) Paragraphs(low, high int) string {
	paragraphs := loremipsum.Paragraphs(f.rng, f.count(low, high), false)
	return strings.Join(paragraphs, "\n\n")
}

// ParagraphsHTML returns paragraphs wrapped in HTML <p> tags.
//
// The number of paragraphs is a fixed count when low equals high, otherwise a
// random count in the inclusive range [low, high].
func (f *Faker) ParagraphsHTML(low, high int) string {
	return linebreaks(f.Paragraphs(low, high))
}

// letters returns count random upper-case ASCII letters.
func (f *Faker) letters(count int) string {
	out := make([]byte, count)
	for i := range out {
		out[i] = byte('A' + f.rng.IntN(26))
	}
	return string(out)
}

// digits returns a string of count decimal digits, never all zeros.
func (f *Faker) digits(count int) string {
	if count < 1 || count > 256 {
		panic("fake: number of digits should be in range 1 to 256")
	}
	for {
		out := make([]byte, count)
		allZero := true
		for i := range out {
			d := f.rng.IntN(10)
			if d != 0 {
				allZero = false
			}
			out[i] = byte('0' + d)
		}
		if !allZero {
			return string(out)
		}
	}
}

// linebreaks converts plain text into HTML paragraphs.
//
// Blank-line separated blocks become <p> elements and single newlines become
// <br> tags, mirroring Django's linebreaks filter without auto-escaping.
func linebreaks(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	paragraphs := paragraphBreak.Split(text, -1)
	for i, paragraph := range paragraphs {
		paragraphs[i] = "<p>" + strings.ReplaceAll(paragraph, "\n", "<br>") + "</p>"
	}
	return strings.Join(paragraphs, "\n\n")
}
