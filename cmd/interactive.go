package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/paul/context-gen/generator"
	"github.com/paul/context-gen/ui"
)

// reader is the input source for interactive prompts (fallback mode).
var reader io.Reader = os.Stdin

func newScanner() *bufio.Scanner {
	return bufio.NewScanner(reader)
}

// runInteractive starts the interactive menu loop.
func runInteractive() {
	if !ui.IsTTY() {
		runInteractiveFallback()
		return
	}

	fmt.Println()
	fmt.Println(ui.Banner())

	for {
		menuItems := []ui.MenuItem{
			{Title: "Generate context files", Key: "init"},
			{Title: "Update existing files", Key: "update"},
			{Title: "Preview output", Key: "preview"},
			{Title: "Help", Key: "help"},
			{Title: "Exit", Key: "exit"},
		}
		menu := ui.NewMenuModel("What would you like to do?", menuItems)

		p := tea.NewProgram(menu)
		result, err := p.Run()
		if err != nil {
			// Bubbletea failed — fall back to plain mode
			runInteractiveFallback()
			return
		}

		m := result.(ui.MenuModel)
		if !m.Chosen() {
			fmt.Println()
			return
		}

		switch m.Selected() {
		case 0:
			interactiveInitTUI()
		case 1:
			interactiveUpdateTUI()
		case 2:
			interactivePreviewTUI()
		case 3:
			fmt.Println()
			printUsage()
		case 4:
			fmt.Println()
			fmt.Println(ui.Success.Render("Bye!"))
			return
		}

		fmt.Println()
		fmt.Print(ui.Subtle.Render("Press Enter to continue..."))
		fmt.Scanln()
	}
}

func interactiveInitTUI() {
	dir := promptDirTUI()
	format := promptFormatTUI()
	if format == "" {
		return
	}

	args := []string{"-d", dir, "-f", format}
	if err := runInit(args); err != nil {
		fmt.Println(ui.FormatError(err.Error()))
	}
}

func interactiveUpdateTUI() {
	dir := promptDirTUI()

	args := []string{"-d", dir}
	if err := runUpdate(args); err != nil {
		fmt.Println(ui.FormatError(err.Error()))
	}
}

func interactivePreviewTUI() {
	dir := promptDirTUI()
	format := promptFormatTUI()
	if format == "" {
		return
	}

	args := []string{"-d", dir, "-f", format}
	if err := runPreview(args); err != nil {
		fmt.Println(ui.FormatError(err.Error()))
	}
}

func promptDirTUI() string {
	m := ui.NewDirPromptModel(".")
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "."
	}
	return result.(ui.DirPromptModel).Value()
}

func promptFormatTUI() string {
	m := ui.NewFormatSelectModel()
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return "both"
	}
	fm := result.(ui.FormatSelectModel)
	if !fm.Chosen() {
		return ""
	}
	return fm.SelectedFormat()
}

// runInteractiveFallback runs the old-style non-TUI interactive mode.
func runInteractiveFallback() {
	scanner := newScanner()

	fmt.Println()
	fmt.Println("context-gen " + ui.Version)
	fmt.Println("Generate AI context files for your codebase")

	for {
		fmt.Println()
		fmt.Println("What would you like to do?")
		fmt.Println()
		fmt.Println("  1. Generate context files (first time)")
		fmt.Println("  2. Update existing context files")
		fmt.Println("  3. Preview output")
		fmt.Println("  4. Help")
		fmt.Println("  5. Exit")
		fmt.Println()

		choice := promptChoice(scanner, "> ", 1, 5)
		if choice == -1 {
			fmt.Println()
			return
		}

		switch choice {
		case 1:
			interactiveInit(scanner)
		case 2:
			interactiveUpdate(scanner)
		case 3:
			interactivePreview(scanner)
		case 4:
			fmt.Println()
			printUsage()
		case 5:
			fmt.Println()
			fmt.Println("Bye!")
			return
		}

		fmt.Println()
		fmt.Print("Press Enter to continue...")
		scanner.Scan()
	}
}

func interactiveInit(scanner *bufio.Scanner) {
	dir := promptString(scanner, "Target directory", ".")
	format := promptFormat(scanner)

	args := []string{"-d", dir, "-f", string(format)}
	if err := runInit(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func interactiveUpdate(scanner *bufio.Scanner) {
	dir := promptString(scanner, "Target directory", ".")

	args := []string{"-d", dir}
	if err := runUpdate(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func interactivePreview(scanner *bufio.Scanner) {
	dir := promptString(scanner, "Target directory", ".")
	format := promptFormat(scanner)

	args := []string{"-d", dir, "-f", string(format)}
	if err := runPreview(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func promptFormat(scanner *bufio.Scanner) generator.Format {
	fmt.Println()
	fmt.Println("Select format:")
	fmt.Println()
	fmt.Println("  1. Claude (CLAUDE.md)")
	fmt.Println("  2. Cursor (.cursorrules)")
	fmt.Println("  3. Both")
	fmt.Println()

	choice := promptChoice(scanner, "> ", 1, 3)

	switch choice {
	case 1:
		return generator.FormatClaude
	case 2:
		return generator.FormatCursor
	default:
		return generator.FormatBoth
	}
}

// promptChoice displays a prompt and reads an integer choice in [min, max].
// Returns -1 on EOF.
func promptChoice(scanner *bufio.Scanner, prompt string, min, max int) int {
	for {
		fmt.Print(prompt)
		if !scanner.Scan() {
			return -1
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		n := 0
		for _, c := range input {
			if c < '0' || c > '9' {
				n = -1
				break
			}
			n = n*10 + int(c-'0')
		}

		if n >= min && n <= max {
			return n
		}

		fmt.Printf("  ! Please enter a number between %d and %d\n", min, max)
	}
}

// promptString displays a prompt with an optional default value and reads a string.
func promptString(scanner *bufio.Scanner, prompt, defaultVal string) string {
	fmt.Println()
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", prompt, defaultVal)
	} else {
		fmt.Printf("%s: ", prompt)
	}

	if !scanner.Scan() {
		return defaultVal
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		return defaultVal
	}
	return input
}
