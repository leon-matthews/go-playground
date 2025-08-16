// Interfaces in practice
package main

import (
	"fmt"
)

// 1. Let consumers define interfaces
// 2. Re-use standard interfaces wherever possible
// 3. Keep interfaces small (bigger the interface, weaker the abstraction)
// 4. Compose interfaces together, if needed.
// 5. Don't couple interfaces to particular types or implementations
// 6. Accept interfaces, return concrete types (ie. let consumer decide how to use it)
//    That is, "Be liberal in what you accept, conservative in what you return", but
//    the [error] interface is an exception to this rule.
func main() {

}
