package generator

import (
	"fmt"
	"strings"

	"github.com/paul/context-gen/analyzer"
)

// Format represents the output format type.
type Format string

const (
	FormatClaude Format = "claude"
	FormatCursor Format = "cursor"
	FormatBoth   Format = "both"
)

// Generate creates context file content for the specified format.
func Generate(info *analyzer.ProjectInfo, format Format) map[string]string {
	results := make(map[string]string)

	switch format {
	case FormatClaude:
		results["CLAUDE.md"] = generateClaude(info)
	case FormatCursor:
		results[".cursorrules"] = generateCursor(info)
	case FormatBoth:
		results["CLAUDE.md"] = generateClaude(info)
		results[".cursorrules"] = generateCursor(info)
	}

	return results
}

func generateClaude(info *analyzer.ProjectInfo) string {
	var b strings.Builder

	b.WriteString("# CLAUDE.md\n\n")
	b.WriteString("This file provides guidance to Claude Code when working with this codebase.\n\n")

	writeProjectOverview(&b, info)
	writeTechStack(&b, info)
	writeBuildCommands(&b, info)
	writeStructure(&b, info)
	writeConventions(&b, info)

	return b.String()
}

func generateCursor(info *analyzer.ProjectInfo) string {
	var b strings.Builder

	b.WriteString("# Project Context\n\n")

	writeProjectOverview(&b, info)
	writeTechStack(&b, info)
	writeBuildCommands(&b, info)
	writeStructure(&b, info)
	writeConventions(&b, info)

	return b.String()
}

func writeProjectOverview(b *strings.Builder, info *analyzer.ProjectInfo) {
	b.WriteString("## Project Overview\n\n")

	if len(info.Languages) > 0 {
		primary := info.Languages[0]
		b.WriteString(fmt.Sprintf("Primary language: **%s** (%.0f%% of codebase)\n", primary.Name, primary.Percentage))
	}

	if len(info.Frameworks) > 0 {
		b.WriteString(fmt.Sprintf("Frameworks: %s\n", strings.Join(info.Frameworks, ", ")))
	}

	b.WriteString(fmt.Sprintf("Files: %d | Directories: %d\n", info.Structure.TotalFiles, info.Structure.TotalDirs))
	b.WriteString("\n")
}

func writeTechStack(b *strings.Builder, info *analyzer.ProjectInfo) {
	b.WriteString("## Tech Stack\n\n")

	if len(info.Languages) > 0 {
		b.WriteString("### Languages\n")
		for _, lang := range info.Languages {
			b.WriteString(fmt.Sprintf("- %s (%d files, %.0f%%)\n", lang.Name, lang.FileCount, lang.Percentage))
		}
		b.WriteString("\n")
	}

	if len(info.BuildTools) > 0 {
		b.WriteString(fmt.Sprintf("### Build Tools\n- %s\n\n", strings.Join(info.BuildTools, "\n- ")))
	}

	if len(info.PackageManagers) > 0 {
		b.WriteString("### Package Managers\n")
		for eco, pm := range info.PackageManagers {
			b.WriteString(fmt.Sprintf("- %s: %s\n", eco, pm))
		}
		b.WriteString("\n")
	}

	if len(info.TestTools) > 0 {
		b.WriteString(fmt.Sprintf("### Testing\n- %s\n\n", strings.Join(info.TestTools, "\n- ")))
	}

	if len(info.Linters) > 0 {
		b.WriteString(fmt.Sprintf("### Linting / Formatting\n- %s\n\n", strings.Join(info.Linters, "\n- ")))
	}

	if info.HasDocker {
		b.WriteString("### Infrastructure\n- Docker\n")
	}
	if info.HasCI {
		b.WriteString(fmt.Sprintf("- CI/CD: %s\n", info.CIProvider))
	}
	if info.HasDocker || info.HasCI {
		b.WriteString("\n")
	}
}

func writeBuildCommands(b *strings.Builder, info *analyzer.ProjectInfo) {
	b.WriteString("## Common Commands\n\n")

	// Generate commands based on detected stack
	if hasLang(info, "Go") {
		b.WriteString("```bash\n")
		b.WriteString("go build ./...        # Build\n")
		b.WriteString("go test ./...         # Run tests\n")
		b.WriteString("go vet ./...          # Lint\n")
		b.WriteString("```\n\n")
	}

	if jsPM := info.PackageManagers["js"]; jsPM != "" && hasJSLang(info) {
		run := jsPM + " run"
		b.WriteString("```bash\n")
		b.WriteString(fmt.Sprintf("%s install          # Install dependencies\n", jsPM))
		b.WriteString(fmt.Sprintf("%s dev              # Development server\n", run))
		b.WriteString(fmt.Sprintf("%s build            # Production build\n", run))
		b.WriteString(fmt.Sprintf("%s test             # Run tests\n", run))
		b.WriteString(fmt.Sprintf("%s lint             # Lint\n", run))
		b.WriteString("```\n\n")
	}

	if hasLang(info, "Python") {
		pyPM := info.PackageManagers["python"]
		switch pyPM {
		case "poetry":
			b.WriteString("```bash\n")
			b.WriteString("poetry install       # Install dependencies\n")
			b.WriteString("poetry run pytest    # Run tests\n")
			b.WriteString("poetry run ruff check .  # Lint\n")
			b.WriteString("```\n\n")
		case "pdm":
			b.WriteString("```bash\n")
			b.WriteString("pdm install          # Install dependencies\n")
			b.WriteString("pdm run pytest       # Run tests\n")
			b.WriteString("```\n\n")
		default:
			b.WriteString("```bash\n")
			b.WriteString("pip install -r requirements.txt  # Install\n")
			b.WriteString("pytest                           # Run tests\n")
			b.WriteString("```\n\n")
		}
	}

	if hasLang(info, "Rust") {
		b.WriteString("```bash\n")
		b.WriteString("cargo build          # Build\n")
		b.WriteString("cargo test           # Run tests\n")
		b.WriteString("cargo clippy         # Lint\n")
		b.WriteString("cargo fmt            # Format\n")
		b.WriteString("```\n\n")
	}
}

func writeStructure(b *strings.Builder, info *analyzer.ProjectInfo) {
	if len(info.Structure.TopLevelDirs) == 0 {
		return
	}

	b.WriteString("## Project Structure\n\n")
	b.WriteString("```\n")
	for _, d := range info.Structure.TopLevelDirs {
		b.WriteString(fmt.Sprintf("%s/\n", d))
	}
	b.WriteString("```\n\n")

	if len(info.Structure.EntryPoints) > 0 {
		b.WriteString("### Entry Points\n")
		for _, e := range info.Structure.EntryPoints {
			b.WriteString(fmt.Sprintf("- `%s`\n", e))
		}
		b.WriteString("\n")
	}
}

func writeConventions(b *strings.Builder, info *analyzer.ProjectInfo) {
	if len(info.Conventions) == 0 {
		return
	}

	b.WriteString("## Conventions\n\n")
	for _, c := range info.Conventions {
		if c.Confidence < 0.5 {
			continue
		}
		b.WriteString(fmt.Sprintf("- **%s**: %s\n", c.Category, c.Description))
		for _, ex := range c.Examples {
			b.WriteString(fmt.Sprintf("  - Example: `%s`\n", ex))
		}
	}
	b.WriteString("\n")
}

func hasLang(info *analyzer.ProjectInfo, name string) bool {
	for _, l := range info.Languages {
		if l.Name == name {
			return true
		}
	}
	return false
}

func hasJSLang(info *analyzer.ProjectInfo) bool {
	return hasLang(info, "JavaScript") || hasLang(info, "TypeScript") ||
		hasLang(info, "TypeScript (React)") || hasLang(info, "JavaScript (React)")
}
