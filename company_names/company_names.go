package main

import (
	"bufio"
	"log"
	"os"
)

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

// Readlines build slice of every non-blank, not-comment line.
// Comments start at the '#' character and continue to the end of the line.
// Whitespace is trimmed from both ends of returned lines.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	lines := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines, nil
}
