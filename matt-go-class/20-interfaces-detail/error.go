package main

import (
	"fmt"
)

func main() {
	// The nil interface behaviour matters for errors, as error is an interface:
	// type error interface {
	//     func Error() string
	// }

	// This works as expected
	err := run(17) // concrete type *myError
	if err == nil {
		fmt.Println("err is nil")
	} else {
		fmt.Println("err is NOT nil")
	}

	// But this more idiomatic usage does NOT!
	// The interface type err2 now has a valid concrete type, so will never
	// be nil - even if the value is.
	var err2 error = run(42) // interface type error
	if err2 == nil {
		fmt.Println("err2 is nil")
	} else {
		fmt.Println("err2 is NOT nil")
	}
}

// It is a MISTAKE to return a concrete type here, because when assigned to
// an [errors.error] interface, that interface will have a concrete type - so
// won't be nil.
func run(a int) *myError {
	return nil
}

// myError implements error
type myError struct {
	err  error
	path string
}

func (e myError) Error() string {
	return fmt.Sprintf("%s: %s", e.path, e.err)
}
