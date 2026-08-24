package main

import (
	"fmt"
	"uuid"
)

func main() {
	fmt.Println("UUID example")
	var u uuid.UUID

	u = uuid.New()
	fmt.Println("New()  ", u)

	u = uuid.NewV4()
	fmt.Println("NewV4()", u)

	u = uuid.NewV7()
	fmt.Println("NewV7()", u)
	u = uuid.NewV7()
	fmt.Println("NewV7()", u)
}
