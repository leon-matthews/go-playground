package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/dustin/go-humanize"

	"local.dev/mimicry"
)

// Extra lines reserved below the header when the detail pane is open: a group summary and a footer.
const detailChrome = 2

// dupItem adapts a DuplicateGroup to the list.DefaultItem interface.
type dupItem struct {
	group mimicry.DuplicateGroup
}

// Title is the shared base name of the duplicated files.
func (i dupItem) Title() string {
	return filepath.Base(i.group.Files[0].Path)
}

// Description is the copy count, per-file size, and reclaimable space of the group.
func (i dupItem) Description() string {
	return fmt.Sprintf(
		"%d copies · %s each · %s reclaimable",
		len(i.group.Files),
		humanize.IBytes(uint64(i.group.Size)),
		humanize.IBytes(uint64(i.group.Reclaimable())),
	)
}

// FilterValue filters on the full path so a directory or name substring both match.
func (i dupItem) FilterValue() string {
	return i.group.Files[0].Path
}

// duplicatesModel is the master/detail browser: a filterable list of groups whose selected
// entry expands into a scrollable pane of the files it contains.
type duplicatesModel struct {
	list        list.Model
	viewport    viewport.Model
	summary     mimicry.Summary
	groupCount  int
	reclaimable int64
	showing     *mimicry.DuplicateGroup // non-nil while the detail pane is open
	ready       bool
	width       int
	height      int
}

// newDuplicatesModel builds the browser over groups (already ordered by reclaimable space).
func newDuplicatesModel(groups []mimicry.DuplicateGroup, summary mimicry.Summary) duplicatesModel {
	items := make([]list.Item, len(groups))
	var reclaimable int64
	for i, g := range groups {
		items[i] = dupItem{group: g}
		reclaimable += g.Reclaimable()
	}
	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Duplicate groups"
	return duplicatesModel{
		list:        l,
		viewport:    viewport.New(),
		summary:     summary,
		groupCount:  len(groups),
		reclaimable: reclaimable,
	}
}

// Init implements tea.Model; the browser needs no startup command.
func (m duplicatesModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m duplicatesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		m.list.SetSize(m.width, m.height-headerHeight)
		m.viewport.SetWidth(m.width)
		m.viewport.SetHeight(m.height - headerHeight - detailChrome)
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	if m.showing != nil {
		m.viewport, cmd = m.viewport.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

// handleKey routes key presses, keeping the filter input and the two panes from fighting.
func (m duplicatesModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// While the filter is being typed, every key belongs to it.
	if m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	if msg.String() == "ctrl+c" || msg.String() == "q" {
		return m, tea.Quit
	}

	if m.showing != nil {
		switch msg.String() {
		case "esc", "left", "h", "backspace":
			m.showing = nil
			return m, nil
		}
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "enter", "right", "l":
		if it, ok := m.list.SelectedItem().(dupItem); ok {
			g := it.group
			m.showing = &g
			m.viewport.SetContent(detailContent(g))
			m.viewport.SetYOffset(0)
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m duplicatesModel) View() tea.View {
	if !m.ready {
		return altView("")
	}
	if m.showing != nil {
		return altView(m.detailView())
	}
	return altView(m.header() + "\n" + m.list.View())
}

// header renders the title and the scan totals shared by both views.
func (m duplicatesModel) header() string {
	stats := fmt.Sprintf(
		"%s files · %s total · %d duplicate groups · %s reclaimable",
		humanize.Comma(int64(m.summary.Count)),
		humanize.IBytes(uint64(m.summary.Size)),
		m.groupCount,
		humanize.IBytes(uint64(m.reclaimable)),
	)
	return titleStyle.Render("mimicry") + "\n" + subtitleStyle.Render(stats)
}

// detailView renders the header, the selected group's summary, its file list, and a footer.
func (m duplicatesModel) detailView() string {
	g := *m.showing
	summary := subtitleStyle.Render(fmt.Sprintf(
		"%s — %d copies, %s each, %s reclaimable",
		filepath.Base(g.Files[0].Path),
		len(g.Files),
		humanize.IBytes(uint64(g.Size)),
		humanize.IBytes(uint64(g.Reclaimable())),
	))
	footer := helpStyle.Render(fmt.Sprintf("%3.0f%% · ↑/↓ scroll · esc back · q quit",
		m.viewport.ScrollPercent()*100))
	return strings.Join([]string{m.header(), summary, m.viewport.View(), footer}, "\n")
}

// detailContent is the scrollable body of the detail pane: every copy's path and modification time.
func detailContent(g mimicry.DuplicateGroup) string {
	var b strings.Builder
	for _, f := range g.Files {
		fmt.Fprintf(&b, "%s\n    %s\n", f.Path, mtimeStyle.Render(f.ModTime.Local().Format("2006-01-02 15:04:05")))
	}
	return b.String()
}
