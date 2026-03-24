package tui

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"banking/categorise"
	"banking/common"
)

// TreeRow is a flattened node in the category tree for display.
type TreeRow struct {
	Name  string // display name, e.g. "Cafe"
	Depth int    // indentation level
	Path  string // full category path, e.g. "Food/Cafe"
}

// BuildCategoryTree builds a sorted, flattened tree from category strings.
func BuildCategoryTree(prefixes []categorise.Prefix) []TreeRow {
	// Collect unique categories and sort them so the tree is stable.
	seen := make(map[string]struct{})
	var categories []string
	for _, p := range prefixes {
		if _, ok := seen[p.Category]; !ok {
			seen[p.Category] = struct{}{}
			categories = append(categories, p.Category)
		}
	}
	slices.Sort(categories)

	// Build a tree of nodes, then flatten.
	type node struct {
		name     string
		children []*node
	}
	var roots []*node

	findChild := func(nodes []*node, name string) *node {
		for _, n := range nodes {
			if n.name == name {
				return n
			}
		}
		return nil
	}

	for _, cat := range categories {
		parts := strings.Split(cat, "/")
		children := &roots
		for _, part := range parts {
			child := findChild(*children, part)
			if child == nil {
				child = &node{name: part}
				*children = append(*children, child)
			}
			children = &child.children
		}
	}

	// Flatten depth-first.
	var rows []TreeRow
	var flatten func(nodes []*node, depth int, prefix string)
	flatten = func(nodes []*node, depth int, prefix string) {
		for _, n := range nodes {
			path := n.name
			if prefix != "" {
				path = prefix + "/" + n.name
			}
			rows = append(rows, TreeRow{Name: n.name, Depth: depth, Path: path})
			flatten(n.children, depth+1, path)
		}
	}
	flatten(roots, 0, "")
	return rows
}

const (
	focusPrefix   = 0
	focusTree     = 1
	focusFreeText = 2

	pageStep = 10
)

// EditorModel is the bubbletea model for interactively categorising unknown transactions.
type EditorModel struct {
	unknowns     []*common.Transaction
	current      int
	prefixInput  textinput.Model
	tree         []TreeRow
	treeCursor   int
	freeText     textinput.Model
	focus        int
	basePrefixes []categorise.Prefix
	Added        []categorise.Prefix
	matcher      *categorise.Matcher
	done         bool
}

// NewEditorModel creates an EditorModel for the given unknown transactions.
func NewEditorModel(unknowns []*common.Transaction, basePrefixes []categorise.Prefix, tree []TreeRow) EditorModel {
	pi := textinput.New()
	pi.Prompt = "  Prefix: "
	pi.Placeholder = "trim to make more general"
	pi.CharLimit = 80
	pi.SetWidth(60)

	ft := textinput.New()
	ft.Prompt = "  Category: "
	ft.Placeholder = "e.g. Food/Cafe"
	ft.CharLimit = 80
	ft.SetWidth(60)

	m := EditorModel{
		unknowns:     unknowns,
		prefixInput:  pi,
		tree:         tree,
		basePrefixes: basePrefixes,
		matcher:      categorise.NewMatcher(basePrefixes),
		freeText:     ft,
	}
	m.prefixInput.SetValue(strings.ToLower(unknowns[0].Details))
	m.prefixInput.CursorEnd()
	m.prefixInput.Focus()

	return m
}

func (m EditorModel) Init() tea.Cmd {
	return m.prefixInput.Focus()
}

func (m EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			m.done = true
			return m, tea.Quit

		case "esc":
			if m.focus == focusFreeText {
				m.focus = focusTree
				m.freeText.Blur()
				m.freeText.Reset()
				return m, nil
			}
			return m.advance()

		case "enter":
			return m.handleEnter()

		case "/":
			if m.focus == focusTree {
				m.focus = focusFreeText
				return m, m.freeText.Focus()
			}
		case "up", "k":
			if m.focus == focusTree && m.treeCursor > 0 {
				m.treeCursor--
				return m, nil
			}
		case "down", "j":
			if m.focus == focusTree && m.treeCursor < len(m.tree)-1 {
				m.treeCursor++
				return m, nil
			}
		case "pgup":
			if m.focus == focusTree {
				m.treeCursor = max(0, m.treeCursor-pageStep)
				return m, nil
			}
		case "pgdown":
			if m.focus == focusTree {
				m.treeCursor = min(len(m.tree)-1, m.treeCursor+pageStep)
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.focus {
	case focusPrefix:
		m.prefixInput, cmd = m.prefixInput.Update(msg)
	case focusFreeText:
		m.freeText, cmd = m.freeText.Update(msg)
	}
	return m, cmd
}

func (m EditorModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.focus {
	case focusPrefix:
		if m.prefixInput.Value() == "" {
			return m.advance()
		}
		if cat := m.matcher.Match(m.prefixInput.Value()); cat != categorise.Unknown {
			return m.saveAndAdvance(cat)
		}
		m.focus = focusTree
		m.prefixInput.Blur()
		return m, nil

	case focusTree:
		if len(m.tree) == 0 {
			return m, nil
		}
		selected := m.tree[m.treeCursor].Path
		return m.saveAndAdvance(selected)

	case focusFreeText:
		if m.freeText.Value() == "" {
			return m, nil
		}
		return m.saveAndAdvance(m.freeText.Value())
	}
	return m, nil
}

func (m EditorModel) saveAndAdvance(category string) (tea.Model, tea.Cmd) {
	m.Added = append(m.Added, categorise.Prefix{
		Text:     strings.ToLower(m.prefixInput.Value()),
		Category: category,
	})
	all := slices.Concat(m.basePrefixes, m.Added)
	m.matcher = categorise.NewMatcher(all)
	m.tree = BuildCategoryTree(all)
	m.treeCursor = 0
	return m.advance()
}

func (m EditorModel) advance() (tea.Model, tea.Cmd) {
	for m.current++; m.current < len(m.unknowns); m.current++ {
		if m.matcher.Match(m.unknowns[m.current].Details) != categorise.Unknown {
			continue
		}
		break
	}
	if m.current >= len(m.unknowns) {
		m.done = true
		return m, tea.Quit
	}
	m.focus = focusPrefix
	m.treeCursor = 0
	m.prefixInput.Reset()
	m.prefixInput.SetValue(strings.ToLower(m.unknowns[m.current].Details))
	m.prefixInput.CursorEnd()
	m.freeText.Reset()
	return m, m.prefixInput.Focus()
}

func (m EditorModel) View() tea.View {
	if m.done {
		return tea.NewView("")
	}

	t := m.unknowns[m.current]
	s := fmt.Sprintf("\n[%d/%d]\n", m.current+1, len(m.unknowns))
	s += fmt.Sprintf("  Date:    %s\n", t.Date.Format(common.DateFormat))
	s += fmt.Sprintf("  Account: %s\n", t.Account)
	s += fmt.Sprintf("  Details: %s\n", t.Details)
	s += fmt.Sprintf("  Amount:  $%.2f\n\n", t.Amount)
	s += m.prefixInput.View() + "\n"

	if m.focus >= focusTree {
		s += "\n"
		if m.focus == focusFreeText {
			s += m.freeText.View() + "\n"
		} else {
			s += m.viewTree()
		}
		s += "\n  Enter select | / new category | Esc back | Ctrl+C quit\n"
	} else {
		s += "\n  Enter confirm | Esc skip | Ctrl+C quit\n"
	}

	return tea.NewView(s)
}

func (m EditorModel) viewTree() string {
	var b strings.Builder
	for i, row := range m.tree {
		indent := strings.Repeat("  ", row.Depth+1)
		cursor := " "
		if i == m.treeCursor {
			cursor = ">"
		}
		fmt.Fprintf(&b, "%s%s %s\n", indent, cursor, row.Name)
	}
	return b.String()
}
