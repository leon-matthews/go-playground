package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"

	"banking/categorise"
)

const (
	catModeBrowse = iota
	catModeRename
	catModeDeleteConfirm
	catModeMove
	catModeMerge
)

// CategoriesModel is the bubbletea model for editing the category tree.
type CategoriesModel struct {
	Prefixes  []categorise.Prefix
	tree      []TreeRow
	cursor    int
	mode      int
	source    int // tree index of the node being moved/merged
	textInput textinput.Model
	Changed   bool
	done      bool
	statusMsg string
}

// NewCategoriesModel creates a CategoriesModel for the given prefixes.
func NewCategoriesModel(prefixes []categorise.Prefix) CategoriesModel {
	ti := textinput.New()
	ti.Prompt = "  New name: "
	ti.CharLimit = 80
	ti.SetWidth(60)

	return CategoriesModel{
		Prefixes:  prefixes,
		tree:      BuildCategoryTree(prefixes),
		textInput: ti,
	}
}

func (m CategoriesModel) Init() tea.Cmd {
	return nil
}

func (m CategoriesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch m.mode {
		case catModeBrowse:
			return m.updateBrowse(msg)
		case catModeRename:
			return m.updateRename(msg)
		case catModeDeleteConfirm:
			return m.updateDeleteConfirm(msg)
		case catModeMove:
			return m.updateMove(msg)
		case catModeMerge:
			return m.updateMerge(msg)
		}
	}
	return m, nil
}

func (m CategoriesModel) updateBrowse(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.done = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.tree)-1 {
			m.cursor++
		}
	case "r":
		if len(m.tree) == 0 {
			return m, nil
		}
		m.mode = catModeRename
		m.textInput.SetValue(m.tree[m.cursor].Name)
		m.textInput.CursorEnd()
		return m, m.textInput.Focus()
	case "d":
		if len(m.tree) == 0 {
			return m, nil
		}
		m.mode = catModeDeleteConfirm
	case "m":
		if len(m.tree) == 0 {
			return m, nil
		}
		m.mode = catModeMove
		m.source = m.cursor
		m.statusMsg = fmt.Sprintf("Moving %q — select new parent", m.tree[m.cursor].Path)
	case "M":
		if len(m.tree) == 0 {
			return m, nil
		}
		m.mode = catModeMerge
		m.source = m.cursor
		m.statusMsg = fmt.Sprintf("Merging %q — select target", m.tree[m.cursor].Path)
	}
	return m, nil
}

func (m CategoriesModel) updateRename(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		newName := strings.TrimSpace(m.textInput.Value())
		if newName != "" && newName != m.tree[m.cursor].Name {
			old := m.tree[m.cursor]
			newPath := newName
			if idx := strings.LastIndex(old.Path, "/"); idx >= 0 {
				newPath = old.Path[:idx+1] + newName
			}
			m.rewritePrefixes(old.Path, newPath)
			m.statusMsg = fmt.Sprintf("Renamed %q → %q", old.Path, newPath)
		}
		m.mode = catModeBrowse
		m.textInput.Blur()
		return m, nil
	case "esc":
		m.mode = catModeBrowse
		m.textInput.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m CategoriesModel) updateDeleteConfirm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		path := m.tree[m.cursor].Path
		m.Prefixes = m.deletePrefixes(path)
		m.tree = BuildCategoryTree(m.Prefixes)
		if m.cursor >= len(m.tree) {
			m.cursor = max(0, len(m.tree)-1)
		}
		m.Changed = true
		m.statusMsg = fmt.Sprintf("Deleted %q", path)
	default:
		m.statusMsg = ""
	}
	m.mode = catModeBrowse
	return m, nil
}

func (m CategoriesModel) updateMove(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = catModeBrowse
		m.statusMsg = ""
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.tree)-1 {
			m.cursor++
		}
	case "enter":
		src := m.tree[m.source]
		dst := m.tree[m.cursor]

		if m.cursor == m.source || strings.HasPrefix(dst.Path, src.Path+"/") {
			m.statusMsg = "Cannot move a category under itself"
			return m, nil
		}

		srcName := src.Name
		newPath := dst.Path + "/" + srcName
		m.rewritePrefixes(src.Path, newPath)
		m.statusMsg = fmt.Sprintf("Moved %q → %q", src.Path, newPath)
	}
	return m, nil
}

func (m CategoriesModel) updateMerge(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = catModeBrowse
		m.statusMsg = ""
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.tree)-1 {
			m.cursor++
		}
	case "enter":
		src := m.tree[m.source]
		dst := m.tree[m.cursor]

		if m.cursor == m.source {
			m.statusMsg = "Cannot merge a category into itself"
			return m, nil
		}

		m.rewritePrefixes(src.Path, dst.Path)
		m.statusMsg = fmt.Sprintf("Merged %q → %q", src.Path, dst.Path)
	}
	return m, nil
}

func (m *CategoriesModel) rewritePrefixes(oldPath, newPath string) {
	for i := range m.Prefixes {
		cat := m.Prefixes[i].Category
		if cat == oldPath {
			m.Prefixes[i].Category = newPath
		} else if strings.HasPrefix(cat, oldPath+"/") {
			m.Prefixes[i].Category = newPath + cat[len(oldPath):]
		}
	}
	m.tree = BuildCategoryTree(m.Prefixes)
	if m.cursor >= len(m.tree) {
		m.cursor = max(0, len(m.tree)-1)
	}
	m.Changed = true
	m.mode = catModeBrowse
}

func (m CategoriesModel) deletePrefixes(path string) []categorise.Prefix {
	var keep []categorise.Prefix
	for _, p := range m.Prefixes {
		if p.Category != path && !strings.HasPrefix(p.Category, path+"/") {
			keep = append(keep, p)
		}
	}
	return keep
}

func (m CategoriesModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}

	var s string
	s += "\n  Category Editor\n\n"
	s += m.viewCatTree()

	if m.mode == catModeRename {
		s += "\n" + m.textInput.View() + "\n"
	} else if m.mode == catModeDeleteConfirm {
		s += fmt.Sprintf("\n  Delete %q and all its prefixes? (y/n)\n", m.tree[m.cursor].Path)
	}

	if m.statusMsg != "" {
		s += "\n  " + m.statusMsg + "\n"
	}

	s += "\n" + m.viewHelp()

	return tea.NewView(s)
}

func (m CategoriesModel) viewCatTree() string {
	var b strings.Builder
	for i, row := range m.tree {
		indent := strings.Repeat("  ", row.Depth+1)

		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		marker := ""
		if (m.mode == catModeMove || m.mode == catModeMerge) && i == m.source {
			marker = " *"
		}

		fmt.Fprintf(&b, "%s%s %s%s\n", indent, cursor, row.Name, marker)
	}
	return b.String()
}

func (m CategoriesModel) viewHelp() string {
	switch m.mode {
	case catModeRename:
		return "  Enter confirm | Esc cancel\n"
	case catModeDeleteConfirm:
		return ""
	case catModeMove:
		return "  Enter move here | Esc cancel | ↑↓ navigate\n"
	case catModeMerge:
		return "  Enter merge into this | Esc cancel | ↑↓ navigate\n"
	default:
		return "  r rename | d delete | m move | M merge | q quit\n"
	}
}
