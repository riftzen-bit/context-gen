package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/paul/context-gen/generator"
)

// reader is the input source for interactive prompts. Defaults to os.Stdin.
// Can be replaced in tests.
var reader io.Reader = os.Stdin

func newScanner() *bufio.Scanner {
	return bufio.NewScanner(reader)
}

// runInteractive starts the interactive menu loop.
func runInteractive() {
	scanner := newScanner()

	fmt.Println()
	fmt.Println(bold("context-gen") + " v0.1.0")
	fmt.Println("Generate AI context files for your codebase")

	for {
		fmt.Println()
		fmt.Println(bold("What would you like to do?"))
		fmt.Println()
		fmt.Println("  1. Generate context files (first time)")
		fmt.Println("  2. Update existing context files")
		fmt.Println("  3. Preview output")
		fmt.Println("  4. Help")
		fmt.Println("  5. Exit")
		fmt.Println()

		choice := promptChoice(scanner, "> ", 1, 5)
		if choice == -1 {
			// EOF (e.g. Ctrl+D)
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
			fmt.Println(green("Bye!"))
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
		errorf("%v", err)
	}
}

func interactiveUpdate(scanner *bufio.Scanner) {
	dir := promptString(scanner, "Target directory", ".")

	args := []string{"-d", dir}
	if err := runUpdate(args); err != nil {
		errorf("%v", err)
	}
}

func interactivePreview(scanner *bufio.Scanner) {
	dir := promptString(scanner, "Target directory", ".")
	format := promptFormat(scanner)

	args := []string{"-d", dir, "-f", string(format)}
	if err := runPreview(args); err != nil {
		errorf("%v", err)
	}
}

func promptFormat(scanner *bufio.Scanner) generator.Format {
	fmt.Println()
	fmt.Println(bold("Select format:"))
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

		fmt.Printf("  %s Please enter a number between %d and %d\n", red("!"), min, max)
	}
}

// promptString displays a prompt with an optional default value and reads a string.
// Returns the default value if the user presses Enter without typing anything.
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
