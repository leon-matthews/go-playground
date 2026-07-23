package humanise

import "strings"

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
