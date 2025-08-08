package main

import (
	"fmt"
)

// Using a struct makes the zero value ready-to-use (and encapsulates implementation)
type StringStack struct {
	data []string	// Not exported
}

func (s *StringStack) Push(x string) {
	s.data = append(s.data, x)
}

func (s *StringStack) Pop() string {
	if l := len(s.data); l > 0 {
		t := s.data[l-1]
		s.data = s.data[:l-1]
		return t
	}
	panic("pop from empty stack")
}

func main() {
	var s StringStack
	s.Push("Leon")
	fmt.Println(s.Pop())
}
