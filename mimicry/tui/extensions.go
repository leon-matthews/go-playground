package tui

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/dustin/go-humanize"

	"local.dev/mimicry"
)

// extItem adapts an ExtensionStat to the list.DefaultItem interface.
type extItem struct {
	stat mimicry.ExtensionStat
}

// Title is the extension, or "(none)" for files without one.
func (i extItem) Title() string {
	if i.stat.Extension == "" {
		return "(none)"
	}
	return i.stat.Extension
}

// Description is the file count and combined size for the extension.
func (i extItem) Description() string {
	return fmt.Sprintf("%d files · %s", i.stat.Count, humanize.IBytes(uint64(i.stat.Size)))
}

// FilterValue filters on the extension name.
func (i extItem) FilterValue() string {
	return i.stat.Extension
}

// extensionsModel is the per-extension browser: a single filterable list, no detail pane.
type extensionsModel struct {
	list     list.Model
	summary  mimicry.Summary
	extCount int
	ready    bool
}

// newExtensionsModel builds the browser over stats (already ordered by descending count).
func newExtensionsModel(stats []mimicry.ExtensionStat, summary mimicry.Summary) extensionsModel {
	items := make([]list.Item, len(stats))
	for i, s := range stats {
		items[i] = extItem{stat: s}
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Extensions"
	return extensionsModel{list: l, summary: summary, extCount: len(stats)}
}

// Init implements tea.Model; the browser needs no startup command.
func (m extensionsModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m extensionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ready = true
		m.list.SetSize(msg.Width, msg.Height-headerHeight)
		return m, nil
	case tea.KeyPressMsg:
		if m.list.FilterState() != list.Filtering {
			if msg.String() == "ctrl+c" || msg.String() == "q" {
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m extensionsModel) View() tea.View {
	if !m.ready {
		return altView("")
	}
	return altView(m.header() + "\n" + m.list.View())
}

// header renders the title and the scan totals.
func (m extensionsModel) header() string {
	stats := fmt.Sprintf(
		"%s files · %s total · %d extensions",
		humanize.Comma(int64(m.summary.Count)),
		humanize.IBytes(uint64(m.summary.Size)),
		m.extCount,
	)
	return titleStyle.Render("mimicry") + "\n" + subtitleStyle.Render(stats)
}
