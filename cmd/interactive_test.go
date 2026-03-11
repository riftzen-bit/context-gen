package cmd

import (
	"bufio"
	"strings"
	"testing"
)

func TestPromptChoice_ValidInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		min   int
		max   int
		want  int
	}{
		{"first option", "1\n", 1, 5, 1},
		{"last option", "5\n", 1, 5, 5},
		{"middle option", "3\n", 1, 5, 3},
		{"single digit range", "2\n", 1, 3, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			got := promptChoice(scanner, "> ", tt.min, tt.max)
			if got != tt.want {
				t.Errorf("promptChoice() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestPromptChoice_InvalidThenValid(t *testing.T) {
	// First "9" is out of range, "abc" is not a number, "3" is valid
	input := "9\nabc\n3\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	got := promptChoice(scanner, "> ", 1, 5)
	if got != 3 {
		t.Errorf("promptChoice() = %d, want 3", got)
	}
}

func TestPromptChoice_EOF(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader(""))
	got := promptChoice(scanner, "> ", 1, 5)
	if got != -1 {
		t.Errorf("promptChoice() on EOF = %d, want -1", got)
	}
}

func TestPromptChoice_EmptyLineRetries(t *testing.T) {
	input := "\n\n2\n"
	scanner := bufio.NewScanner(strings.NewReader(input))
	got := promptChoice(scanner, "> ", 1, 5)
	if got != 2 {
		t.Errorf("promptChoice() = %d, want 2", got)
	}
}

func TestPromptString_WithInput(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("./my-project\n"))
	got := promptString(scanner, "Target directory", ".")
	if got != "./my-project" {
		t.Errorf("promptString() = %q, want %q", got, "./my-project")
	}
}

func TestPromptString_EmptyUsesDefault(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("\n"))
	got := promptString(scanner, "Target directory", ".")
	if got != "." {
		t.Errorf("promptString() = %q, want %q", got, ".")
	}
}

func TestPromptString_EOFUsesDefault(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader(""))
	got := promptString(scanner, "Target directory", "/home")
	if got != "/home" {
		t.Errorf("promptString() = %q, want %q", got, "/home")
	}
}

func TestPromptString_WhitespaceOnlyUsesDefault(t *testing.T) {
	scanner := bufio.NewScanner(strings.NewReader("   \n"))
	got := promptString(scanner, "Target directory", ".")
	if got != "." {
		t.Errorf("promptString() = %q, want %q", got, ".")
	}
}

func TestPromptFormat(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"claude", "1\n", "claude"},
		{"cursor", "2\n", "cursor"},
		{"both", "3\n", "both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(strings.NewReader(tt.input))
			got := promptFormat(scanner)
			if string(got) != tt.want {
				t.Errorf("promptFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}
