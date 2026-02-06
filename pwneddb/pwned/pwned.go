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
	"fmt"
	"iter"
)

// Prefixes are currently 5 hexadecimal characters long
const prefixLength = 5

// A Prefix is a hexadecimal string used to reference a [HashList], eg. "cafe5"
type Prefix string

// NewPrefix casts a string to a Prefix, checking its validity
func NewPrefix(prefix string) (Prefix, error) {
	if len(prefix) != prefixLength {
		return "", fmt.Errorf("prefix must contain %d characters: %q", prefixLength, prefix)
	}
	return Prefix(prefix), nil
}

// A HashList is a multi-line string containing password hashes and counts.
// Specifically, 2 colon-separated values: HEX(SHA1(password)):count
// eg. "308672AB94BCBE0B2FEE2EC68FC69F9D5E6:8"
type HashList string

// Prefixes generates all possible hexadecimal prefixes
func Prefixes() iter.Seq[Prefix] {
	limit := 0x01 << (prefixLength * 4)
	return func(yield func(Prefix) bool) {
		for v := range limit {
			hex := fmt.Sprintf("%0*x", prefixLength, v)
			if !yield(Prefix(hex)) {
				return
			}
		}
	}
}
