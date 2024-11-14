package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTable(t *testing.T) {
	table := NewTable([]string{"Length", "Count"})

	t.Run("append row", func(t *testing.T) {
		assert.Equal(t, 0, table.NumRows())
		table.AppendRow([]string{"35 -> 39", "36"})
		err := table.AppendRow([]string{"40 -> 44", "8"})
		assert.Nil(t, err)
		assert.Equal(t, 2, table.NumRows())
	})

	t.Run("append row error", func(t *testing.T) {
		assert.Equal(t, 2, table.NumRows())
		err := table.AppendRow([]string{"too", "many", "rows"})
		assert.Equal(t, "expected 2 columns, got 3", err.Error())
	})

	// TODO
	t.Run("print table", func(t *testing.T) {
		fmt.Println(table.Print())
	})
}
