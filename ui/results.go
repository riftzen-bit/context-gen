package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/paul/context-gen/analyzer"
)

// FormatResults formats detection results as a styled box.
func FormatResults(info *analyzer.ProjectInfo) string {
	var lines []string

	if len(info.Languages) > 0 {
		var langParts []string
		for _, l := range info.Languages {
			langParts = append(langParts, fmt.Sprintf("%s (%.0f%%)", l.Name, l.Percentage))
		}
		lines = append(lines, fmt.Sprintf("%-14s%s", Bold.Render("Languages"), strings.Join(langParts, ", ")))
	}

	if len(info.Frameworks) > 0 {
		lines = append(lines, fmt.Sprintf("%-14s%s", Bold.Render("Frameworks"), strings.Join(info.Frameworks, ", ")))
	}

	if len(info.BuildTools) > 0 {
		lines = append(lines, fmt.Sprintf("%-14s%s", Bold.Render("Build"), strings.Join(info.BuildTools, ", ")))
	}

	if len(info.PackageManagers) > 0 {
		var pms []string
		keys := make([]string, 0, len(info.PackageManagers))
		for k := range info.PackageManagers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, eco := range keys {
			pms = append(pms, fmt.Sprintf("%s:%s", eco, info.PackageManagers[eco]))
		}
		lines = append(lines, fmt.Sprintf("%-14s%s", Bold.Render("Packages"), strings.Join(pms, ", ")))
	}

	if len(info.TestTools) > 0 {
		lines = append(lines, fmt.Sprintf("%-14s%s", Bold.Render("Testing"), strings.Join(info.TestTools, ", ")))
	}

	if len(info.Linters) > 0 {
		lines = append(lines, fmt.Sprintf("%-14s%s", Bold.Render("Linting"), strings.Join(info.Linters, ", ")))
	}

	if info.HasDocker {
		lines = append(lines, fmt.Sprintf("%-14s%s", Bold.Render("Docker"), "yes"))
	}
	if info.HasCI {
		lines = append(lines, fmt.Sprintf("%-14s%s", Bold.Render("CI"), info.CIProvider))
	}

	content := strings.Join(lines, "\n")
	return ResultsStyle.Render(content)
}
