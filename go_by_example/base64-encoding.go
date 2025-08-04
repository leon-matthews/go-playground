package main

import (
	"encoding/base64"
	"fmt"
)

func main() {
	data := "abc123!?$*&()'-=@~"

	// Standard Base64 (RFC 4648)
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	fmt.Println(encoded) // YWJjMTIzIT8kKiYoKSctPUB+
	decoded, _ := base64.StdEncoding.DecodeString(encoded)
	fmt.Println(string(decoded))

	// URL-safe Base64 (also RFC 4648)
	encoded = base64.URLEncoding.EncodeToString([]byte(data))
	fmt.Println(encoded) // YWJjMTIzIT8kKiYoKSctPUB-
	decoded, _ = base64.URLEncoding.DecodeString(encoded)
	fmt.Println(string(decoded))
}
