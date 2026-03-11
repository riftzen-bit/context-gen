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
	"Both",
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

// DirPromptModel is a bubbletea model for directory input.
type DirPromptModel struct {
	input    string
	defVal   string
	done     bool
	quitting bool
}

// NewDirPromptModel creates a directory prompt with a default value.
func NewDirPromptModel(defaultVal string) DirPromptModel {
	return DirPromptModel{
		defVal: defaultVal,
	}
}

// Value returns the entered directory or default.
func (m DirPromptModel) Value() string {
	if m.input == "" {
		return m.defVal
	}
	return m.input
}

func (m DirPromptModel) Init() tea.Cmd {
	return nil
}

func (m DirPromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			m.done = true
			return m, tea.Quit
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}
	}

	return m, nil
}

func (m DirPromptModel) View() string {
	prompt := "\n" + Bold.Render("Target directory") + " [" + m.defVal + "]: "
	cursor := "█"
	return prompt + m.input + cursor + "\n"
}
