package iteration

import "strings"

const count = 5

func Repeat(s string) string {
	sb := strings.Builder{}
	for range count {
		sb.WriteString(s)
	}
	return sb.String()
}
