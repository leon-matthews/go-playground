package lib

import (
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"unicode"
)

// ToAscii strips accents and other combining glyphs from Unicode input
func ToAscii(str string) (string, error) {
	// Unicode category 'Mn' is Mark, nonspacing
	result, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn))), str)
	if err != nil {
		return "", err
	}
	return result, nil
}

// TODO Try smaz compression
//  https://github.com/cespare/go-smaz
//var b = smaz.Compress([]byte(line))
