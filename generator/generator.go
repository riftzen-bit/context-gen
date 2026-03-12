package generator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/riftzen-bit/context-gen/analyzer"
	"github.com/riftzen-bit/context-gen/ui"
)

// Format represents the output format type.
type Format string

const (
	FormatClaude    Format = "claude"
	FormatCursor    Format = "cursor"
	FormatBoth      Format = "both"
	FormatAgents    Format = "agents"
	FormatCursorMDC Format = "cursor-mdc"
	FormatCline     Format = "cline"
	FormatWindsurf      Format = "windsurf"
	FormatAntigravity   Format = "antigravity"
	FormatAll           Format = "all"
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
	case FormatAgents:
		results["AGENTS.md"] = generateAgents(info)
	case FormatCursorMDC:
		results[".cursor/rules/project.mdc"] = generateCursorMDC(info)
	case FormatCline:
		results[".clinerules"] = generateCline(info)
	case FormatWindsurf:
		results[".windsurfrules"] = generateWindsurf(info)
	case FormatAntigravity:
		results[".gemini/GEMINI.md"] = generateAntigravity(info)
	case FormatAll:
		results["CLAUDE.md"] = generateClaude(info)
		results[".cursorrules"] = generateCursor(info)
		results["AGENTS.md"] = generateAgents(info)
		results[".cursor/rules/project.mdc"] = generateCursorMDC(info)
		results[".clinerules"] = generateCline(info)
		results[".windsurfrules"] = generateWindsurf(info)
		results[".gemini/GEMINI.md"] = generateAntigravity(info)
	}

	return results
}

func generateClaude(info *analyzer.ProjectInfo) string {
	var b strings.Builder

	writeHeader(&b, info, true)
	writeAgentBody(&b, info)

	return b.String()
}

// writeAgentBody writes the shared body used by agent-first formats (Claude, Antigravity).
// Includes all standard sections plus Workflow/Architecture placeholders.
func writeAgentBody(b *strings.Builder, info *analyzer.ProjectInfo) {
	writeTechStackConcise(b, info)
	writeCommands(b, info)
	writeCodeStyle(b, info)
	writeStructureAnnotated(b, info)
	writeKeyFiles(b, info)
	writeConventions(b, info)
	writeRules(b, info)
	writeTesting(b, info)
	writePlaceholder(b, "Workflow", "Add git workflow, branch naming, PR conventions here")
	writePlaceholder(b, "Architecture", "Add key design decisions here")
	writeFooter(b)
}

func generateCursor(info *analyzer.ProjectInfo) string {
	var b strings.Builder

	writeHeader(&b, info, false)
	writeTechStackConcise(&b, info)
	writeCommands(&b, info)
	writeCodeStyle(&b, info)
	writeStructureAnnotated(&b, info)
	writeKeyFiles(&b, info)
	writeConventions(&b, info)
	writeRules(&b, info)
	writeTesting(&b, info)
	writeFooter(&b)

	return b.String()
}

func generateAgents(info *analyzer.ProjectInfo) string {
	var b strings.Builder

	name := info.Name
	if name == "" {
		name = "Project"
	}
	b.WriteString(fmt.Sprintf("# AGENTS.md - %s\n\n", name))
	if info.Description != "" {
		b.WriteString(info.Description + "\n\n")
	}

	writeTechStackConcise(&b, info)
	writeCommands(&b, info)
	writeCodeStyle(&b, info)
	writeStructureAnnotated(&b, info)
	writeKeyFiles(&b, info)
	writeConventions(&b, info)
	writeRules(&b, info)
	writeTesting(&b, info)
	writeFooter(&b)

	return b.String()
}

func generateCursorMDC(info *analyzer.ProjectInfo) string {
	var b strings.Builder

	name := info.Name
	if name == "" {
		name = "Project"
	}
	b.WriteString("---\n")
	b.WriteString(fmt.Sprintf("description: Project context and coding guidelines for %s\n", name))
	b.WriteString("globs: \n")
	b.WriteString("alwaysApply: true\n")
	b.WriteString("---\n\n")

	writeHeader(&b, info, false)
	writeTechStackConcise(&b, info)
	writeCommands(&b, info)
	writeCodeStyle(&b, info)
	writeStructureAnnotated(&b, info)
	writeKeyFiles(&b, info)
	writeConventions(&b, info)
	writeRules(&b, info)
	writeTesting(&b, info)
	writeFooter(&b)

	return b.String()
}

func generateCline(info *analyzer.ProjectInfo) string {
	return generateCursor(info)
}

func generateWindsurf(info *analyzer.ProjectInfo) string {
	return generateCursor(info)
}

func generateAntigravity(info *analyzer.ProjectInfo) string {
	var b strings.Builder

	name := info.Name
	if name == "" {
		name = "Project"
	}
	b.WriteString(fmt.Sprintf("# GEMINI.md - %s\n\n", name))
	if info.Description != "" {
		b.WriteString(info.Description + "\n\n")
	}

	writeAgentBody(&b, info)

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
	for _, lang := range info.Languages {
		parts = append(parts, lang.Name)
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
	if len(info.Linters) > 0 {
		b.WriteString("- Linters: " + strings.Join(info.Linters, ", ") + "\n")
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
		ann := ""
		if a, ok := dirAnnotations[d]; ok {
			ann = " — " + a
		}
		b.WriteString(fmt.Sprintf("%-12s%s\n", d+"/", ann))

		// Show subdirs for key directories
		if info.Structure.SubDirs != nil {
			if subs, ok := info.Structure.SubDirs[d]; ok && len(subs) > 0 {
				sorted := make([]string, len(subs))
				copy(sorted, subs)
				sort.Strings(sorted)
				for _, sub := range sorted {
					b.WriteString(fmt.Sprintf("  %-10s\n", sub+"/"))
				}
			}
		}
	}
	b.WriteString("```\n\n")
}

func writeKeyFiles(b *strings.Builder, info *analyzer.ProjectInfo) {
	if len(info.Structure.EntryPoints) == 0 {
		return
	}
	b.WriteString("## Key Files\n\n")
	for _, ep := range info.Structure.EntryPoints {
		b.WriteString(fmt.Sprintf("- `%s`\n", ep))
	}
	b.WriteString("\n")
}

func writePlaceholder(b *strings.Builder, section, hint string) {
	b.WriteString(fmt.Sprintf("## %s\n\n", section))
	b.WriteString(fmt.Sprintf("<!-- %s -->\n\n", hint))
}

func writeTesting(b *strings.Builder, info *analyzer.ProjectInfo) {
	b.WriteString("## Testing\n\n")

	if len(info.TestTools) > 0 {
		b.WriteString("Framework: " + strings.Join(info.TestTools, ", ") + "\n\n")
	}

	b.WriteString("```bash\n")
	wrote := false
	if hasLang(info, "Go") {
		b.WriteString("go test ./...                          # Run all tests\n")
		b.WriteString("go test ./pkg/name -run TestFunc       # Run single test\n")
		b.WriteString("go test ./... -v                       # Verbose output\n")
		wrote = true
	}
	if hasJSLang(info) {
		if wrote {
			b.WriteString("\n")
		}
		pm := info.PackageManagers["js"]
		if pm == "" {
			pm = "npm"
		}
		run := pm + " run"
		b.WriteString(fmt.Sprintf("%-42s # Run all tests\n", run+" test"))
		b.WriteString(fmt.Sprintf("%-42s # Run single file\n", run+" test -- path/to/file.test.ts"))
		wrote = true
	}
	if hasLang(info, "Python") {
		if wrote {
			b.WriteString("\n")
		}
		b.WriteString("pytest                                 # Run all tests\n")
		b.WriteString("pytest path/to/test_file.py::test_name # Run single test\n")
		b.WriteString("pytest -x                              # Stop on first failure\n")
		wrote = true
	}
	if hasLang(info, "Rust") {
		if wrote {
			b.WriteString("\n")
		}
		b.WriteString("cargo test                             # Run all tests\n")
		b.WriteString("cargo test test_name                   # Run single test\n")
		b.WriteString("cargo test -- --nocapture              # Show output\n")
		wrote = true
	}
	if !wrote {
		b.WriteString("# Add test commands here\n")
	}
	b.WriteString("```\n\n")
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

func writeRules(b *strings.Builder, info *analyzer.ProjectInfo) {
	rules := RelevantRules(info)
	if len(rules) == 0 {
		return
	}

	b.WriteString("## Coding Guidelines\n\n")
	for _, rule := range rules {
		b.WriteString(fmt.Sprintf("- %s\n", rule))
	}
	b.WriteString("\n")
}

func writeFooter(b *strings.Builder) {
	b.WriteString(fmt.Sprintf("<!-- Generated by context-gen %s -->\n", ui.Version))
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
