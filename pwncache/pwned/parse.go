package pwned

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

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
	hashes := make([]Hash, 0, strings.Count(string(list), "\n")+1)
	number := 0
	for line := range strings.Lines(string(list)) {
		number++
		line = strings.TrimRight(line, "\r\n")
		// Tolerate a trailing newline on the last line
		if line == "" {
			continue
		}

		suffix, countField, found := strings.Cut(line, ":")
		if !found {
			return nil, fmt.Errorf("hash list line %d: no colon separator", number)
		}
		if len(suffix) != suffixLength {
			return nil, fmt.Errorf(
				"hash list line %d: suffix must contain %d characters: %q",
				number, suffixLength, suffix,
			)
		}

		sha1, err := hex.DecodeString(string(prefix) + suffix)
		if err != nil {
			return nil, fmt.Errorf("hash list line %d: %w", number, err)
		}
		count, err := strconv.ParseInt(countField, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("hash list line %d: %w", number, err)
		}
		hashes = append(hashes, Hash{SHA1: sha1, Count: count})
	}
	return hashes, nil
}
