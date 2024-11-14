package main

import (
	"bufio"
	"os"
	"strings"
)

type Name []byte

func NewName(name string) Name {
	return Name(name)
}

func (n Name) Length() int {
	return len(n)
}

func (n Name) String() string {
	return string(n)
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

// ShortestAndLongest finds the lengths of the shortest and longest names
func ShortestAndLongest(names []Name) (shortest int, longest int) {
	if len(names) == 0 {
		return 0, 0
	}

	shortest = names[0].Length()
	longest = shortest
	for _, name := range names {
		shortest = min(shortest, name.Length())
		longest = max(longest, name.Length())
	}
	return shortest, longest
}
