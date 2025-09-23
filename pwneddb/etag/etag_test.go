package etag_test

import (
	"fmt"
	"testing"

	"pwneddb/etag"
)

func TestETagStore(t *testing.T) {
	records := map[string]string{
		"cafe5": `W/"0x8DDD9FCAE5BEBD4"`,
	}

	store := etag.NewETagStore()
	store["cafe5"] = `W/"0x8DDD9FCAE5BEBD4"`
	fmt.Printf("[%T]%+[1]v\n", records)
	store.Save("")
}
