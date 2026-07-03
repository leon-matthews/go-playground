package pwned

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
)

// SHA-1 hashes are 20 bytes, or half their hexadecimal length
const sha1Length = hashHexLength / 2

// A Hash is a single entry parsed from a [HashList]
type Hash struct {
	SHA1  []byte // Full 20-byte SHA-1 of the password
	Count int64  // Times the password has appeared in a breach
}

// ParseHashList converts a hash list into full hashes and their counts.
// Each line carries a 35-character hexadecimal suffix and a count, which the
// prefix completes to a full 40-character hash. Lines end with CRLF.
func ParseHashList(prefix Prefix, list HashList) ([]Hash, error) {
	const suffixLength = hashHexLength - prefixLength

	// Newlines bound the line count, sizing both slices so neither grows
	// mid-parse. The bound also keeps every backing reslice within cap, so all
	// of a list's hashes share one allocation rather than one slice per line.
	estimate := bytes.Count(list, []byte("\n")) + 1
	hashes := make([]Hash, 0, estimate)
	backing := make([]byte, 0, estimate*sha1Length)

	// scratch holds one full 40-character hash. The prefix never changes, so
	// only the per-line suffix is copied in before each decode.
	var scratch [hashHexLength]byte
	copy(scratch[:prefixLength], prefix)

	number := 0
	for line := range bytes.Lines(list) {
		number++
		line = bytes.TrimRight(line, "\r\n")
		// Tolerate a trailing newline on the last line
		if len(line) == 0 {
			continue
		}

		suffix, countField, found := bytes.Cut(line, []byte(":"))
		if !found {
			return nil, fmt.Errorf("hash list line %d: no colon separator", number)
		}
		if len(suffix) != suffixLength {
			return nil, fmt.Errorf(
				"hash list line %d: suffix must contain %d characters: %q",
				number, suffixLength, suffix,
			)
		}

		// Carve the next 20-byte window from backing and decode straight into it
		copy(scratch[prefixLength:], suffix)
		start := len(backing)
		backing = backing[:start+sha1Length]
		sha1 := backing[start : start+sha1Length : start+sha1Length]
		if _, err := hex.Decode(sha1, scratch[:]); err != nil {
			return nil, fmt.Errorf("hash list line %d: %w", number, err)
		}
		count, err := parseCount(countField)
		if err != nil {
			return nil, fmt.Errorf("hash list line %d: %w", number, err)
		}
		hashes = append(hashes, Hash{SHA1: sha1, Count: count})
	}
	return hashes, nil
}

// parseCount reads a non-negative base-10 count straight from field.
// Working from the bytes avoids the per-line string allocation that
// strconv.ParseInt would need, while still rejecting overflow.
func parseCount(field []byte) (int64, error) {
	if len(field) == 0 {
		return 0, fmt.Errorf("empty count")
	}
	var count int64
	for _, c := range field {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid count %q", field)
		}
		digit := int64(c - '0')
		if count > (math.MaxInt64-digit)/10 {
			return 0, fmt.Errorf("count %q out of range", field)
		}
		count = count*10 + digit
	}
	return count, nil
}
