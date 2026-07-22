// Package tui is the interactive terminal browser for a mimicry report.
package tui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"local.dev/mimicry"
)

// Shared styles for both browsers; kept minimal so they render on any colour profile.
var (
	titleStyle    = lipgloss.NewStyle().Bold(true)
	subtitleStyle = lipgloss.NewStyle().Faint(true)
	helpStyle     = lipgloss.NewStyle().Faint(true)
	mtimeStyle    = lipgloss.NewStyle().Faint(true)
)

// Header chrome reserved above the active component: a title line and a stats line.
const headerHeight = 2

// RunDuplicates launches the duplicate-group browser and blocks until the user quits.
func RunDuplicates(groups []mimicry.DuplicateGroup, summary mimicry.Summary) error {
	_, err := tea.NewProgram(newDuplicatesModel(groups, summary)).Run()
	return err
}

// RunExtensions launches the per-extension browser and blocks until the user quits.
func RunExtensions(stats []mimicry.ExtensionStat, summary mimicry.Summary) error {
	_, err := tea.NewProgram(newExtensionsModel(stats, summary)).Run()
	return err
}

// altView wraps rendered content in a full-screen view that restores the terminal on exit.
func altView(content string) tea.View {
	v := tea.NewView(content)
	v.AltScreen = true
	v.WindowTitle = "mimicry"
	return v
}
