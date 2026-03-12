package ui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/term"
	"github.com/mattn/go-isatty"
)

const Version = "v0.4.0"

var (
	// Colors
	ColorPrimary   = lipgloss.Color("39")  // blue
	ColorSuccess   = lipgloss.Color("82")  // green
	ColorWarning   = lipgloss.Color("214") // yellow
	ColorError     = lipgloss.Color("196") // red
	ColorSubtle    = lipgloss.Color("241") // gray
	ColorHighlight = lipgloss.Color("212") // pink

	// Text styles
	Bold    = lipgloss.NewStyle().Bold(true)
	Success = lipgloss.NewStyle().Foreground(ColorSuccess)
	Warning = lipgloss.NewStyle().Foreground(ColorWarning)
	Error   = lipgloss.NewStyle().Foreground(ColorError)
	Subtle  = lipgloss.NewStyle().Foreground(ColorSubtle)

	// Banner style
	BannerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 2)

	// Results box style
	ResultsStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSubtle).
			Padding(0, 1)

	// Menu item styles
	MenuSelected   = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	MenuUnselected = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	MenuCursor     = lipgloss.NewStyle().Foreground(ColorPrimary).SetString("> ")
	MenuNoCursor   = lipgloss.NewStyle().SetString("  ")
)

// Banner returns the styled welcome banner.
func Banner() string {
	content := Bold.Render("context-gen") + " " + Version + "\n" +
		Subtle.Render("Generate AI context files")
	return BannerStyle.Render(content)
}

// IsTTY returns true if stdout is a terminal.
func IsTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// TermWidth returns the terminal width, defaulting to 80 if detection fails.
func TermWidth() int {
	w, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// TermHeight returns the terminal height, defaulting to 24 if detection fails.
func TermHeight() int {
	_, h, err := term.GetSize(os.Stdout.Fd())
	if err != nil || h <= 0 {
		return 24
	}
	return h
}

// FormatFileCreated formats a file creation message.
func FormatFileCreated(path string) string {
	return "  " + Success.Render("CREATE") + " " + path
}

// FormatFileUpdated formats a file update message.
func FormatFileUpdated(path string) string {
	return "  " + Success.Render("UPDATE") + " " + path
}

// FormatFileSkipped formats a file skip message.
func FormatFileSkipped(name, reason string) string {
	return "  " + Warning.Render("SKIP") + " " + name + " " + reason
}

// FormatError formats an error message.
func FormatError(msg string) string {
	return Error.Render("Error: ") + msg
}
