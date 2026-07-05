// Package pwned downloads and maintains a local copy of the Have I Been Pwned password database.
// https://haveibeenpwned.com/Passwords
//
// Downloading the entire database is free, and does not require an access key,
// so please use good manners. Over a million HTTP requests are required!
//
// The database does not contain actual users' passwords. Instead, it stores
// the SHA1 cryptographic hash of each password, plus the count of times that
// password has appeared in a security breach. These hash lists are grouped by
// a 5-character hexadecimal prefix - which is why a million requests are
// needed, as 16^5 = 1,048,576.
//
// A mapping between prefixes and HTTP ETags is maintained so that a local
// database can be efficiently updated, downloading only hash lists that have
// been changed since the last update.
package pwned

import (
	"encoding/hex"
	"fmt"
	"iter"
	"strconv"
	"strings"
)

// Prefixes are currently 5 hexadecimal characters long
const prefixLength = 5

// PrefixCount is the number of possible prefixes, and therefore hash lists
const PrefixCount = 1 << (prefixLength * 4)

// Full SHA-1 hashes are 40 hexadecimal characters long
const hashHexLength = 40

// A Prefix is a hexadecimal string used to reference a [HashList], eg. "cafe5"
type Prefix string

// NewPrefix casts a string to a Prefix, checking its validity
func NewPrefix(prefix string) (Prefix, error) {
	if len(prefix) != prefixLength {
		return "", fmt.Errorf("prefix must contain %d characters: %q", prefixLength, prefix)
	}
	return Prefix(prefix), nil
}

// HashRange returns the inclusive bounds of all possible hashes sharing this prefix.
// The bounds are full 20-byte hashes, suitable for BETWEEN queries on the database.
func (p Prefix) HashRange() (lower, upper []byte, err error) {
	const fill = hashHexLength - prefixLength
	lower, err = hex.DecodeString(string(p) + strings.Repeat("0", fill))
	if err != nil {
		return nil, nil, fmt.Errorf("prefix %q is not hexadecimal: %w", p, err)
	}
	upper, err = hex.DecodeString(string(p) + strings.Repeat("f", fill))
	if err != nil {
		return nil, nil, fmt.Errorf("prefix %q is not hexadecimal: %w", p, err)
	}
	return lower, upper, nil
}

// Index returns the prefix's position in the dense keyspace, from 0 to PrefixCount-1.
// A valid prefix always lands in range, so it is safe to index a slice sized to
// PrefixCount.
func (p Prefix) Index() (int, error) {
	if len(p) != prefixLength {
		return 0, fmt.Errorf("prefix must contain %d characters: %q", prefixLength, string(p))
	}
	value, err := strconv.ParseUint(string(p), 16, 32)
	if err != nil {
		return 0, fmt.Errorf("prefix %q is not hexadecimal: %w", p, err)
	}
	return int(value), nil
}

// A HashList is the raw bytes of a multi-line hash list.
// Each line carries 2 colon-separated values: HEX(SHA1(password)):count
// eg. "308672AB94BCBE0B2FEE2EC68FC69F9D5E6:8"
type HashList []byte

// Prefixes generates all possible hexadecimal prefixes
func Prefixes() iter.Seq[Prefix] {
	return func(yield func(Prefix) bool) {
		for v := range PrefixCount {
			hex := fmt.Sprintf("%0*x", prefixLength, v)
			if !yield(Prefix(hex)) {
				return
			}
		}
	}
}
