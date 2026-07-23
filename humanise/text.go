package humanise

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// And joins items into an English list with the Oxford comma, eg. "apples, oranges, and pears".
//
// One item returns that item alone; two items are joined with "and" and no comma;
// three or more take the serial comma. An empty or nil slice returns "" and items
// are joined verbatim, without trimming or dropping blanks.
func And(items []string) string {
	return oxford(items, "and")
}

// Or joins items into an English list with the Oxford comma, eg. "apples, oranges, or pears".
//
// One item returns that item alone; two items are joined with "or" and no comma;
// three or more take the serial comma. An empty or nil slice returns "" and items
// are joined verbatim, without trimming or dropping blanks.
func Or(items []string) string {
	return oxford(items, "or")
}

// oxford joins items into a list, placing conjunction before the final item.
func oxford(items []string, conjunction string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " " + conjunction + " " + items[1]
	default:
		last := len(items) - 1
		return strings.Join(items[:last], ", ") + ", " + conjunction + " " + items[last]
	}
}

// titleMinorWords are the short words Title leaves lowercase between the first and last words.
var titleMinorWords = map[string]struct{}{
	"a": {}, "an": {}, "and": {}, "but": {}, "by": {}, "for": {},
	"from": {}, "in": {}, "of": {}, "the": {}, "with": {},
}

// Title capitalises a string as a title, eg. "taming of the shrew" becomes "Taming of the Shrew".
//
// Each word's first letter is capitalised and the rest is left as typed, so
// deliberate capitals such as "McDonald" and "NASA" survive. A short set of minor
// words (a, an, and, but, by, for, from, in, of, the, with) is lowercased between the
// first and last words, which are always capitalised. Whitespace runs collapse to
// single spaces.
func Title(title string) string {
	words := strings.Fields(title)
	last := len(words) - 1
	for i, word := range words {
		lower := strings.ToLower(word)
		if _, minor := titleMinorWords[lower]; i > 0 && i < last && minor {
			words[i] = lower
		} else {
			words[i] = capitaliseFirst(word)
		}
	}
	return strings.Join(words, " ")
}

// capitaliseFirst upper-cases the first rune of word, leaving the rest unchanged.
func capitaliseFirst(word string) string {
	r, size := utf8.DecodeRuneInString(word)
	if r == utf8.RuneError {
		return word
	}
	return string(unicode.ToTitle(r)) + word[size:]
}
