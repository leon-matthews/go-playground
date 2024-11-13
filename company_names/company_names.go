package main

import (
	"bufio"
	"fmt"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"math"
	"os"
	"strings"
	"unicode"
)

type Name struct {
	utf8 []byte
}

func NewName(name string) Name {
	return Name{utf8: []byte(name)}
}

func (n Name) Length() int {
	return len(n.utf8)
}

func (n Name) String() string {
	return string(n.utf8)
}

// CountLengths builds a map of name lengths vs count
func CountLengths(names []Name) map[int]int {
	counts := make(map[int]int)
	for _, name := range names {
		counts[name.Length()]++
	}
	return counts
}

// printHistogram groups counts into buckets of binSize and prints basic histogram
func printHistogram(counts map[int]int, binSize, largest int) {
	numBins := int(math.Ceil(float64(largest) / float64(binSize)))
	for bin := 0; bin < numBins; bin++ {
		count := 0
		for i := 0; i < binSize; i++ {
			index := binSize*bin + i
			c, ok := counts[index]
			if ok {
				count += c
			}
		}
		rangeStart := bin * binSize
		rangeEnd := rangeStart + (binSize - 1)

		fmt.Printf("%2d -> %2d: %v\n", rangeStart, rangeEnd, strings.Repeat("#", count))
	}
}

// ReadNames builds a slice of names from every non-blank, non-comment line.
// Whitespace is trimmed from both ends of input lines, while comments are
// lines that start with the '#' character.
func ReadNames(path string) ([]Name, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	names := make([]Name, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var str string
		str = scanner.Text()
		str = strings.TrimSpace(str)
		if str == "" || str[0] == '#' {
			continue
		}
		names = append(names, NewName(str))
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return names, nil
}

// ShortAndTall finds the lengths of the shortest and longest lines
func ShortAndTall(lines []string) (int, int) {
	if len(lines) == 0 {
		return 0, 0
	}

	shortest := len(lines[0])
	longest := shortest
	for _, line := range lines {
		shortest = min(shortest, len(line))
		longest = max(longest, len(line))
	}
	return shortest, longest
}

func ToAscii(str string) (string, error) {
	result, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn))), str)
	if err != nil {
		return "", err
	}
	return result, nil
}
