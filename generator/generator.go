package generator

import (
	"fmt"
	"sort"
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

// Directory annotations for Project Structure section.
var dirAnnotations = map[string]string{
	"cmd":      "CLI entry points",
	"pkg":      "Public library code",
	"internal": "Private application code",
	"api":      "API definitions",
	"web":      "Web assets",
	"src":      "Source code",
	"lib":      "Library code",
	"test":     "Tests",
	"tests":    "Tests",
	"docs":     "Documentation",
}

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

	writeHeader(&b, info, true)
	writeTechStackConcise(&b, info)
	writeCommands(&b, info)
	writeCodeStyle(&b, info)
	writeStructureAnnotated(&b, info)
	writeConventions(&b, info)
	writePlaceholder(&b, "Workflow", "Add git workflow, branch naming, PR conventions here")
	writePlaceholder(&b, "Architecture", "Add key design decisions here")

	return b.String()
}

func generateCursor(info *analyzer.ProjectInfo) string {
	var b strings.Builder

	writeHeader(&b, info, false)
	writeTechStackConcise(&b, info)
	writeCommands(&b, info)
	writeCodeStyle(&b, info)
	writeStructureAnnotated(&b, info)
	writeConventions(&b, info)
	// Cursor: no Workflow/Architecture placeholders

	return b.String()
}

func writeHeader(b *strings.Builder, info *analyzer.ProjectInfo, isClaude bool) {
	if isClaude {
		name := info.Name
		if name == "" {
			name = "Project"
		}
		b.WriteString(fmt.Sprintf("# %s\n\n", name))
	} else {
		b.WriteString("# Project Context\n\n")
	}

	if info.Description != "" {
		b.WriteString(info.Description + "\n\n")
	}
}

func writeTechStackConcise(b *strings.Builder, info *analyzer.ProjectInfo) {
	b.WriteString("## Tech Stack\n\n")

	var parts []string
	if len(info.Languages) > 0 {
		parts = append(parts, info.Languages[0].Name)
	}
	for _, fw := range info.Frameworks {
		parts = append(parts, fw)
	}
	if info.HasDocker {
		parts = append(parts, "Docker")
	}
	if info.HasCI {
		parts = append(parts, info.CIProvider)
	}

	if len(parts) > 0 {
		b.WriteString(strings.Join(parts, " | "))
		b.WriteString("\n\n")
	}
}

func writeCommands(b *strings.Builder, info *analyzer.ProjectInfo) {
	b.WriteString("## Common Commands\n\n")

	// If detected scripts exist, use them (priority over hardcoded)
	if len(info.Scripts) > 0 {
		b.WriteString("```bash\n")
		keys := make([]string, 0, len(info.Scripts))
		for k := range info.Scripts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, name := range keys {
			b.WriteString(fmt.Sprintf("%-25s # %s\n", info.Scripts[name], name))
		}
		b.WriteString("```\n\n")
		return
	}

	// Fallback: hardcoded language-specific commands
	writeBuildCommands(b, info)
}

func writeCodeStyle(b *strings.Builder, info *analyzer.ProjectInfo) {
	b.WriteString("## Code Style\n\n")

	hasContent := false
	if info.CodeStyle.IndentStyle != "" {
		b.WriteString(fmt.Sprintf("- Indent: %s", info.CodeStyle.IndentStyle))
		if info.CodeStyle.IndentSize > 0 {
			b.WriteString(fmt.Sprintf(" (%d)", info.CodeStyle.IndentSize))
		}
		b.WriteString("\n")
		hasContent = true
	}
	if info.CodeStyle.LineLength > 0 {
		b.WriteString(fmt.Sprintf("- Line length: %d\n", info.CodeStyle.LineLength))
		hasContent = true
	}
	if info.CodeStyle.Formatter != "" {
		b.WriteString(fmt.Sprintf("- Formatter: %s\n", info.CodeStyle.Formatter))
		hasContent = true
	}
	for _, rule := range info.CodeStyle.ExtraRules {
		b.WriteString(fmt.Sprintf("- %s\n", rule))
		hasContent = true
	}

	if !hasContent {
		b.WriteString("<!-- Add project-specific style rules here -->\n")
	}
	b.WriteString("\n")
}

func writeStructureAnnotated(b *strings.Builder, info *analyzer.ProjectInfo) {
	if len(info.Structure.TopLevelDirs) == 0 {
		return
	}

	b.WriteString("## Project Structure\n\n")
	b.WriteString("```\n")
	for _, d := range info.Structure.TopLevelDirs {
		if ann, ok := dirAnnotations[d]; ok {
			b.WriteString(fmt.Sprintf("%-12s — %s\n", d+"/", ann))
		} else {
			b.WriteString(d + "/\n")
		}
	}
	b.WriteString("```\n\n")
}

func writePlaceholder(b *strings.Builder, section, hint string) {
	b.WriteString(fmt.Sprintf("## %s\n\n", section))
	b.WriteString(fmt.Sprintf("<!-- %s -->\n\n", hint))
}

func writeBuildCommands(b *strings.Builder, info *analyzer.ProjectInfo) {
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
