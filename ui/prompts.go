package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// FormatSelectModel is a bubbletea model for selecting output format.
type FormatSelectModel struct {
	cursor   int
	chosen   bool
	quitting bool
}

var formatOptions = []string{
	"Claude (CLAUDE.md)",
	"Cursor (.cursorrules)",
	"GitHub Copilot (AGENTS.md)",
	"Cursor MDC (.cursor/rules/)",
	"Cline (.clinerules)",
	"Windsurf (.windsurfrules)",
	"Antigravity (.gemini/GEMINI.md)",
	"Both (Claude + Cursor)",
	"All formats",
}

// NewFormatSelectModel creates a format selection prompt.
func NewFormatSelectModel() FormatSelectModel {
	return FormatSelectModel{}
}

// SelectedFormat returns the chosen format string.
func (m FormatSelectModel) SelectedFormat() string {
	switch m.cursor {
	case 0:
		return "claude"
	case 1:
		return "cursor"
	case 2:
		return "agents"
	case 3:
		return "cursor-mdc"
	case 4:
		return "cline"
	case 5:
		return "windsurf"
	case 6:
		return "antigravity"
	case 7:
		return "both"
	case 8:
		return "all"
	default:
		return "both"
	}
}

// Chosen returns true if user made a selection.
func (m FormatSelectModel) Chosen() bool {
	return m.chosen
}

func (m FormatSelectModel) Init() tea.Cmd {
	return nil
}

func (m FormatSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(formatOptions)-1 {
				m.cursor++
			}
		case "enter":
			m.chosen = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m FormatSelectModel) View() string {
	s := "\n" + Bold.Render("Select format:") + "\n\n"

	for i, opt := range formatOptions {
		cursor := MenuNoCursor.String()
		style := MenuUnselected
		if m.cursor == i {
			cursor = MenuCursor.String()
			style = MenuSelected
		}
		s += fmt.Sprintf("%s%s\n", cursor, style.Render(opt))
	}

	s += "\n" + Subtle.Render("↑/↓ to navigate • enter to select") + "\n"
	return s
}

