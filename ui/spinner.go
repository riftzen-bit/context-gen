package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ScanDoneMsg signals that scanning is complete.
type ScanDoneMsg struct{}

// SpinnerModel wraps bubbles spinner for scanning animation.
type SpinnerModel struct {
	spinner spinner.Model
	message string
	done    bool
}

// NewSpinnerModel creates a spinner with a scanning message.
func NewSpinnerModel(message string) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)
	return SpinnerModel{
		spinner: s,
		message: message,
	}
}

func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case ScanDoneMsg:
		m.done = true
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.(tea.KeyMsg).String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m SpinnerModel) View() string {
	if m.done {
		return Success.Render("✓") + " " + m.message + " done\n"
	}
	return m.spinner.View() + " " + m.message + "...\n"
}
