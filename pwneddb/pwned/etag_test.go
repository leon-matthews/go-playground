package pwned_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pwneddb/pwned"
)

func TestETagStore(t *testing.T) {
	expected := pwned.ETags{
		"cafe5": `W/"0x8DDD9FCAE5BEBD4"`,
		"cafe6": `W/"0x8DE3759E02B938F"`,
		"bad5a": `W/"0x7EA652281224AB6"`,
	}

	t.Run("load", func(t *testing.T) {
		etags, err := pwned.ETagsLoad("testdata/etags.txt")
		require.NoError(t, err)
		assert.Equal(t, expected, etags)
	})

	t.Run("load not found", func(t *testing.T) {
		etags, err := pwned.ETagsLoad("no/such/file.txt")
		assert.ErrorContains(t, err, "loading etags: open no/such/file.txt: no such file or directory")
		assert.Nil(t, etags)
	})

	// store := pwned.NewETagStore()
	// store["cafe5"] = `W/"0x8DDD9FCAE5BEBD4"`
	// fmt.Printf("[%T]%+[1]v\n", expected)
	// store.Save("etags.txt")
}
