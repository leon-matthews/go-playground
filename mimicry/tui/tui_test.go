package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"local.dev/mimicry"
)

func TestDupItem(t *testing.T) {
	g := mimicry.DuplicateGroup{
		Size: 2048,
		Files: []mimicry.FileInfo{
			{Path: "/home/a/report.pdf"},
			{Path: "/home/b/report.pdf"},
			{Path: "/home/c/report.pdf"},
		},
	}
	it := dupItem{group: g}
	assert.Equal(t, "report.pdf", it.Title())
	assert.Equal(t, "3 copies · 2.0 KiB each · 4.0 KiB reclaimable", it.Description())
	assert.Equal(t, "/home/a/report.pdf", it.FilterValue())
}

func TestExtItem(t *testing.T) {
	assert.Equal(t, ".go", extItem{stat: mimicry.ExtensionStat{Extension: ".go"}}.Title())
	assert.Equal(t, "(none)", extItem{stat: mimicry.ExtensionStat{Extension: ""}}.Title())

	it := extItem{stat: mimicry.ExtensionStat{Extension: ".go", Count: 4, Size: 1024}}
	assert.Equal(t, "4 files · 1.0 KiB", it.Description())
	assert.Equal(t, ".go", it.FilterValue())
}

func TestDetailContent(t *testing.T) {
	g := mimicry.DuplicateGroup{Files: []mimicry.FileInfo{
		{Path: "/home/a/report.pdf"},
		{Path: "/home/b/report.pdf"},
	}}
	out := detailContent(g)
	assert.Contains(t, out, "/home/a/report.pdf")
	assert.Contains(t, out, "/home/b/report.pdf")
	assert.Equal(t, 2, strings.Count(out, "report.pdf"))
}

// sized returns a duplicates model that has received an initial window size.
func sized(t *testing.T) duplicatesModel {
	t.Helper()
	groups := []mimicry.DuplicateGroup{{
		Hash: [32]byte{1},
		Size: 1000,
		Files: []mimicry.FileInfo{
			{Path: "/a/dup"},
			{Path: "/b/dup"},
		},
	}}
	m := newDuplicatesModel(groups, mimicry.Summary{Count: 2, Size: 2000})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return next.(duplicatesModel)
}

func TestDuplicatesDrillDownAndBack(t *testing.T) {
	m := sized(t)
	require.Nil(t, m.showing)

	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = next.(duplicatesModel)
	require.NotNil(t, m.showing, "enter should open the detail pane")
	assert.Equal(t, "/a/dup", m.showing.Files[0].Path)

	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m = next.(duplicatesModel)
	assert.Nil(t, m.showing, "esc should close the detail pane")
}

func TestDuplicatesQuit(t *testing.T) {
	m := sized(t)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	require.NotNil(t, cmd)
	assert.IsType(t, tea.QuitMsg{}, cmd())
}
