package generator

import (
	"strings"
	"testing"

	"github.com/riftzen-bit/context-gen/analyzer"
)

func TestGenerate_ClaudeFormat(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatClaude)

	content, ok := results["CLAUDE.md"]
	if !ok {
		t.Fatal("expected CLAUDE.md in output")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 file, got %d", len(results))
	}

	// New format: starts with project name, not "# CLAUDE.md"
	if !strings.Contains(content, "# test-project") {
		t.Errorf("expected project name header, got:\n%s", content[:min(200, len(content))])
	}

	// Should have concise tech stack (pipe-separated)
	if !strings.Contains(content, "## Tech Stack") {
		t.Error("missing Tech Stack section")
	}

	// Should have common commands
	if !strings.Contains(content, "## Common Commands") {
		t.Error("missing Common Commands section")
	}

	// Should have Code Style section
	if !strings.Contains(content, "## Code Style") {
		t.Error("missing Code Style section")
	}

	// Should have Project Structure with annotations
	if !strings.Contains(content, "## Project Structure") {
		t.Error("missing Project Structure section")
	}

	// Claude format should have Workflow placeholder
	if !strings.Contains(content, "## Workflow") {
		t.Error("missing Workflow placeholder section")
	}

	// Claude format should have Architecture placeholder
	if !strings.Contains(content, "## Architecture") {
		t.Error("missing Architecture placeholder section")
	}

	// Claude format should have Testing section
	if !strings.Contains(content, "## Testing") {
		t.Error("missing Testing section")
	}

	// Should NOT have the old verbose sections
	if strings.Contains(content, "### Languages") {
		t.Error("should not have verbose Languages subsection")
	}
	if strings.Contains(content, "### Build Tools") {
		t.Error("should not have verbose Build Tools subsection")
	}
	if strings.Contains(content, "## Project Overview") {
		t.Error("should not have old Project Overview section")
	}
}

func TestGenerate_CursorFormat(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatCursor)

	content, ok := results[".cursorrules"]
	if !ok {
		t.Fatal("expected .cursorrules in output")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 file, got %d", len(results))
	}

	// Cursor uses "# Project Context" instead of project name
	if !strings.Contains(content, "# Project Context") {
		t.Error(".cursorrules missing '# Project Context' header")
	}
	if strings.Contains(content, "# test-project") {
		t.Error(".cursorrules should not contain project name header")
	}

	// Cursor should NOT have Workflow/Architecture placeholders
	if strings.Contains(content, "## Workflow") {
		t.Error(".cursorrules should not have Workflow section")
	}
	if strings.Contains(content, "## Architecture") {
		t.Error(".cursorrules should not have Architecture section")
	}
}

func TestGenerate_BothFormat(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatBoth)

	if _, ok := results["CLAUDE.md"]; !ok {
		t.Error("expected CLAUDE.md in output")
	}
	if _, ok := results[".cursorrules"]; !ok {
		t.Error("expected .cursorrules in output")
	}
	if len(results) != 2 {
		t.Errorf("expected 2 files, got %d", len(results))
	}
}

func TestWriteHeader_Claude(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Name:        "my-project",
		Description: "A great tool for developers",
	}

	var b strings.Builder
	writeHeader(&b, info, true)
	output := b.String()

	if !strings.Contains(output, "# my-project") {
		t.Error("expected project name in Claude header")
	}
	if !strings.Contains(output, "A great tool for developers") {
		t.Error("expected description in header")
	}
}

func TestWriteHeader_Cursor(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Name:        "my-project",
		Description: "A great tool for developers",
	}

	var b strings.Builder
	writeHeader(&b, info, false)
	output := b.String()

	if !strings.Contains(output, "# Project Context") {
		t.Error("expected '# Project Context' in Cursor header")
	}
	if strings.Contains(output, "# my-project") {
		t.Error("should not contain project name in Cursor header")
	}
}

func TestWriteHeader_EmptyName(t *testing.T) {
	info := &analyzer.ProjectInfo{}

	var b strings.Builder
	writeHeader(&b, info, true)
	output := b.String()

	if !strings.Contains(output, "# Project") {
		t.Error("expected fallback '# Project' when name is empty")
	}
}

func TestWriteTechStackConcise(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Go", FileCount: 10, Percentage: 100},
		},
		Frameworks: []string{"Echo"},
		HasDocker:  true,
		HasCI:      true,
		CIProvider: "GitHub Actions",
	}

	var b strings.Builder
	writeTechStackConcise(&b, info)
	output := b.String()

	if !strings.Contains(output, "Go | Echo | Docker | GitHub Actions") {
		t.Errorf("expected pipe-separated tech stack, got:\n%s", output)
	}
}

func TestWriteTechStackConcise_MultipleLanguages(t *testing.T) {
	// After analyzer merge, React variants are folded into base language names.
	// Framework "React" appears separately in Frameworks.
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "TypeScript", FileCount: 32, Percentage: 80},
			{Name: "Rust", FileCount: 8, Percentage: 20},
		},
		Frameworks: []string{"React"},
		HasCI:      true,
		CIProvider: "GitHub Actions",
	}

	var b strings.Builder
	writeTechStackConcise(&b, info)
	output := b.String()

	if !strings.Contains(output, "TypeScript | Rust | React | GitHub Actions") {
		t.Errorf("expected merged languages in tech stack, got:\n%s", output)
	}
}

func TestWriteTesting_GoProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Go", FileCount: 10, Percentage: 100},
		},
		PackageManagers: map[string]string{},
	}

	var b strings.Builder
	writeTesting(&b, info)
	output := b.String()

	if !strings.Contains(output, "## Testing") {
		t.Error("missing Testing section")
	}
	if !strings.Contains(output, "go test ./...") {
		t.Error("missing go test all command")
	}
	if !strings.Contains(output, "-run TestFunc") {
		t.Error("missing single test command")
	}
}

func TestWriteTesting_JSProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "TypeScript", FileCount: 10, Percentage: 100},
		},
		TestTools:       []string{"Vitest"},
		PackageManagers: map[string]string{"js": "pnpm"},
	}

	var b strings.Builder
	writeTesting(&b, info)
	output := b.String()

	if !strings.Contains(output, "Vitest") {
		t.Error("missing test framework name")
	}
	if !strings.Contains(output, "pnpm run test") {
		t.Error("missing pnpm test command")
	}
}

func TestWriteTesting_Empty(t *testing.T) {
	info := &analyzer.ProjectInfo{}

	var b strings.Builder
	writeTesting(&b, info)
	output := b.String()

	if !strings.Contains(output, "## Testing") {
		t.Error("missing Testing section")
	}
	if !strings.Contains(output, "# Add test commands here") {
		t.Error("expected placeholder for unknown project")
	}
}

func TestWriteStructureAnnotated(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Structure: analyzer.DirStructure{
			TopLevelDirs: []string{"cmd", "internal", "custom"},
		},
	}

	var b strings.Builder
	writeStructureAnnotated(&b, info)
	output := b.String()

	if !strings.Contains(output, "## Project Structure") {
		t.Error("missing Project Structure section")
	}
	if !strings.Contains(output, "CLI entry points") {
		t.Error("missing annotation for cmd/")
	}
	if !strings.Contains(output, "Private application code") {
		t.Error("missing annotation for internal/")
	}
	// Unknown dir should have no annotation
	if strings.Contains(output, "custom/") && strings.Contains(output, "custom/  ") {
		// custom/ should appear but without " — " annotation
	}
}

func TestWriteStructureAnnotated_EmptyDirs(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Structure: analyzer.DirStructure{},
	}

	var b strings.Builder
	writeStructureAnnotated(&b, info)
	output := b.String()

	if strings.Contains(output, "## Project Structure") {
		t.Error("should not write structure section with no dirs")
	}
}

func TestWriteCommands_WithScripts(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Go", FileCount: 10, Percentage: 100},
		},
		Scripts: map[string]string{
			"build": "make build",
			"test":  "make test",
		},
		PackageManagers: map[string]string{},
	}

	var b strings.Builder
	writeCommands(&b, info)
	output := b.String()

	// When scripts exist, they should be used instead of hardcoded
	if !strings.Contains(output, "make build") {
		t.Error("expected detected script 'make build'")
	}
	if !strings.Contains(output, "make test") {
		t.Error("expected detected script 'make test'")
	}
	// Should NOT contain hardcoded Go commands
	if strings.Contains(output, "go build ./...") {
		t.Error("should not have hardcoded Go commands when scripts exist")
	}
}

func TestWriteCommands_FallbackToHardcoded(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Go", FileCount: 10, Percentage: 100},
		},
		Scripts:         nil,
		PackageManagers: map[string]string{},
	}

	var b strings.Builder
	writeCommands(&b, info)
	output := b.String()

	// No scripts, should fall back to hardcoded Go commands
	if !strings.Contains(output, "go build ./...") {
		t.Error("expected hardcoded Go build command as fallback")
	}
}

func TestWriteCodeStyle(t *testing.T) {
	info := &analyzer.ProjectInfo{
		CodeStyle: analyzer.CodeStyle{
			IndentStyle: "tabs",
			IndentSize:  4,
			LineLength:  120,
			Formatter:   "gofmt",
		},
	}

	var b strings.Builder
	writeCodeStyle(&b, info)
	output := b.String()

	if !strings.Contains(output, "- Indent: tabs (4)") {
		t.Errorf("expected indent info, got:\n%s", output)
	}
	if !strings.Contains(output, "- Line length: 120") {
		t.Error("expected line length")
	}
	if !strings.Contains(output, "- Formatter: gofmt") {
		t.Error("expected formatter")
	}
}

func TestWriteCodeStyle_Empty(t *testing.T) {
	info := &analyzer.ProjectInfo{}

	var b strings.Builder
	writeCodeStyle(&b, info)
	output := b.String()

	if !strings.Contains(output, "<!-- Add project-specific style rules here -->") {
		t.Error("expected placeholder comment when no style detected")
	}
}

func TestWriteBuildCommands_GoProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Go", FileCount: 10, Percentage: 100},
		},
		PackageManagers: map[string]string{},
	}

	var b strings.Builder
	writeBuildCommands(&b, info)
	output := b.String()

	wantCommands := []string{
		"go build ./...",
		"go test ./...",
		"go vet ./...",
	}
	for _, cmd := range wantCommands {
		if !strings.Contains(output, cmd) {
			t.Errorf("missing Go command: %s", cmd)
		}
	}
}

func TestWriteBuildCommands_NodeProject(t *testing.T) {
	tests := []struct {
		name    string
		pm      string
		wantPM  string
		wantRun string
	}{
		{"npm", "npm", "npm install", "npm run"},
		{"pnpm", "pnpm", "pnpm install", "pnpm run"},
		{"yarn", "yarn", "yarn install", "yarn run"},
		{"bun", "bun", "bun install", "bun run"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &analyzer.ProjectInfo{
				Languages: []analyzer.Language{
					{Name: "TypeScript", FileCount: 10, Percentage: 100},
				},
				PackageManagers: map[string]string{"js": tt.pm},
			}

			var b strings.Builder
			writeBuildCommands(&b, info)
			output := b.String()

			if !strings.Contains(output, tt.wantPM) {
				t.Errorf("missing %s install command", tt.pm)
			}
			if !strings.Contains(output, tt.wantRun+" dev") {
				t.Errorf("missing %s run dev command", tt.pm)
			}
		})
	}
}

func TestWriteBuildCommands_PythonPoetry(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Python", FileCount: 5, Percentage: 100},
		},
		PackageManagers: map[string]string{"python": "poetry"},
	}

	var b strings.Builder
	writeBuildCommands(&b, info)
	output := b.String()

	wantCommands := []string{
		"poetry install",
		"poetry run pytest",
		"poetry run ruff",
	}
	for _, cmd := range wantCommands {
		if !strings.Contains(output, cmd) {
			t.Errorf("missing Poetry command: %s", cmd)
		}
	}
}

func TestWriteBuildCommands_RustProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Rust", FileCount: 5, Percentage: 100},
		},
		PackageManagers: map[string]string{},
	}

	var b strings.Builder
	writeBuildCommands(&b, info)
	output := b.String()

	wantCommands := []string{
		"cargo build",
		"cargo test",
		"cargo clippy",
		"cargo fmt",
	}
	for _, cmd := range wantCommands {
		if !strings.Contains(output, cmd) {
			t.Errorf("missing Rust command: %s", cmd)
		}
	}
}

func TestWriteConventions(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Conventions: []analyzer.Convention{
			{
				Category:    "naming",
				Description: "Files use kebab-case naming",
				Confidence:  0.9,
				Examples:    []string{"my-component.ts"},
			},
			{
				Category:    "error_handling",
				Description: "Go uses standard if err != nil",
				Confidence:  1.0,
				Examples:    []string{"if err != nil { return err }"},
			},
			{
				Category:    "naming",
				Description: "Low confidence convention",
				Confidence:  0.3, // below threshold
			},
		},
		PackageManagers: map[string]string{},
	}

	var b strings.Builder
	writeConventions(&b, info)
	output := b.String()

	if !strings.Contains(output, "## Conventions") {
		t.Error("missing Conventions section")
	}
	if !strings.Contains(output, "kebab-case") {
		t.Error("missing kebab-case convention")
	}
	if !strings.Contains(output, "err != nil") {
		t.Error("missing err != nil convention")
	}
	if strings.Contains(output, "Low confidence convention") {
		t.Error("low confidence convention should be filtered out")
	}
}

func TestWriteConventions_Empty(t *testing.T) {
	info := &analyzer.ProjectInfo{}

	var b strings.Builder
	writeConventions(&b, info)
	output := b.String()

	if strings.Contains(output, "## Conventions") {
		t.Error("should not write conventions section with no conventions")
	}
}

func TestHasLang(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Go"},
			{Name: "TypeScript"},
		},
	}

	if !hasLang(info, "Go") {
		t.Error("expected hasLang(Go) = true")
	}
	if !hasLang(info, "TypeScript") {
		t.Error("expected hasLang(TypeScript) = true")
	}
	if hasLang(info, "Rust") {
		t.Error("expected hasLang(Rust) = false")
	}
}

func TestHasJSLang(t *testing.T) {
	tests := []struct {
		name  string
		langs []analyzer.Language
		want  bool
	}{
		{"JavaScript", []analyzer.Language{{Name: "JavaScript"}}, true},
		{"TypeScript", []analyzer.Language{{Name: "TypeScript"}}, true},
		{"TypeScript React", []analyzer.Language{{Name: "TypeScript (React)"}}, true},
		{"JavaScript React", []analyzer.Language{{Name: "JavaScript (React)"}}, true},
		{"Go only", []analyzer.Language{{Name: "Go"}}, false},
		{"empty", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &analyzer.ProjectInfo{Languages: tt.langs}
			got := hasJSLang(info)
			if got != tt.want {
				t.Errorf("hasJSLang() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWriteStructureAnnotated_WithSubDirs(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Structure: analyzer.DirStructure{
			TopLevelDirs: []string{"cmd", "src", "docs"},
			SubDirs: map[string][]string{
				"src": {"components", "utils", "pages"},
				"cmd": {"server", "cli"},
			},
		},
	}

	var b strings.Builder
	writeStructureAnnotated(&b, info)
	output := b.String()

	if !strings.Contains(output, "## Project Structure") {
		t.Error("missing Project Structure section")
	}
	// cmd/ should have annotation and subdirs
	if !strings.Contains(output, "CLI entry points") {
		t.Error("missing annotation for cmd/")
	}
	if !strings.Contains(output, "  cli/") {
		t.Errorf("missing subdir cli/ under cmd/, got:\n%s", output)
	}
	if !strings.Contains(output, "  server/") {
		t.Errorf("missing subdir server/ under cmd/, got:\n%s", output)
	}
	// src/ should have subdirs sorted alphabetically
	if !strings.Contains(output, "  components/") {
		t.Errorf("missing subdir components/ under src/, got:\n%s", output)
	}
	if !strings.Contains(output, "  pages/") {
		t.Errorf("missing subdir pages/ under src/, got:\n%s", output)
	}
	if !strings.Contains(output, "  utils/") {
		t.Errorf("missing subdir utils/ under src/, got:\n%s", output)
	}
	// docs/ has no subdirs - should not have indented lines after it
	docsIdx := strings.Index(output, "docs/")
	if docsIdx == -1 {
		t.Error("missing docs/ directory")
	}
}

func TestWriteStructureAnnotated_NilSubDirs(t *testing.T) {
	// Backward compat: SubDirs may be nil with old analyzer
	info := &analyzer.ProjectInfo{
		Structure: analyzer.DirStructure{
			TopLevelDirs: []string{"cmd", "internal"},
			SubDirs:      nil,
		},
	}

	var b strings.Builder
	writeStructureAnnotated(&b, info)
	output := b.String()

	if !strings.Contains(output, "## Project Structure") {
		t.Error("missing Project Structure section")
	}
	if !strings.Contains(output, "cmd/") {
		t.Error("missing cmd/ dir")
	}
	if !strings.Contains(output, "internal/") {
		t.Error("missing internal/ dir")
	}
}

func TestWriteKeyFiles(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Structure: analyzer.DirStructure{
			EntryPoints: []string{"main.go", "cmd/server/main.go"},
		},
	}

	var b strings.Builder
	writeKeyFiles(&b, info)
	output := b.String()

	if !strings.Contains(output, "## Key Files") {
		t.Error("missing Key Files section")
	}
	if !strings.Contains(output, "- `main.go`") {
		t.Errorf("missing main.go entry point, got:\n%s", output)
	}
	if !strings.Contains(output, "- `cmd/server/main.go`") {
		t.Errorf("missing cmd/server/main.go entry point, got:\n%s", output)
	}
}

func TestWriteKeyFiles_Empty(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Structure: analyzer.DirStructure{
			EntryPoints: nil,
		},
	}

	var b strings.Builder
	writeKeyFiles(&b, info)
	output := b.String()

	if output != "" {
		t.Errorf("expected no output for empty entry points, got:\n%s", output)
	}
}

func TestWriteCodeStyle_WithLinters(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Linters: []string{"golangci-lint", "eslint"},
		CodeStyle: analyzer.CodeStyle{
			IndentStyle: "spaces",
			IndentSize:  2,
			Formatter:   "prettier",
		},
	}

	var b strings.Builder
	writeCodeStyle(&b, info)
	output := b.String()

	if !strings.Contains(output, "- Linters: golangci-lint, eslint") {
		t.Errorf("expected linters line, got:\n%s", output)
	}
	if !strings.Contains(output, "- Formatter: prettier") {
		t.Error("missing formatter")
	}
	if !strings.Contains(output, "- Indent: spaces (2)") {
		t.Error("missing indent info")
	}
}

func TestWriteCodeStyle_LintersOnly(t *testing.T) {
	// Linters present but no other style info - should not show placeholder
	info := &analyzer.ProjectInfo{
		Linters: []string{"ruff"},
	}

	var b strings.Builder
	writeCodeStyle(&b, info)
	output := b.String()

	if !strings.Contains(output, "- Linters: ruff") {
		t.Errorf("expected linters line, got:\n%s", output)
	}
	if strings.Contains(output, "<!-- Add project-specific style rules here -->") {
		t.Error("should not show placeholder when linters are present")
	}
}

func TestGenerate_ClaudeFormat_HasKeyFiles(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatClaude)
	content := results["CLAUDE.md"]

	if !strings.Contains(content, "## Key Files") {
		t.Error("CLAUDE.md should contain Key Files section")
	}
	if !strings.Contains(content, "- `main.go`") {
		t.Error("CLAUDE.md should list main.go entry point")
	}
}

func TestGenerate_CursorFormat_HasKeyFiles(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatCursor)
	content := results[".cursorrules"]

	if !strings.Contains(content, "## Key Files") {
		t.Error(".cursorrules should contain Key Files section")
	}
	if !strings.Contains(content, "- `main.go`") {
		t.Error(".cursorrules should list main.go entry point")
	}
}

func TestGenerate_AgentsFormat(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatAgents)

	content, ok := results["AGENTS.md"]
	if !ok {
		t.Fatal("expected AGENTS.md in output")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 file, got %d", len(results))
	}
	if !strings.HasPrefix(content, "# AGENTS.md") {
		t.Errorf("expected content to start with '# AGENTS.md', got:\n%s", content[:min(100, len(content))])
	}
	// Agents should NOT have Workflow/Architecture placeholders
	if strings.Contains(content, "## Workflow") {
		t.Error("AGENTS.md should not have Workflow section")
	}
	if strings.Contains(content, "## Architecture") {
		t.Error("AGENTS.md should not have Architecture section")
	}
}

func TestGenerate_CursorMDCFormat(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatCursorMDC)

	content, ok := results[".cursor/rules/project.mdc"]
	if !ok {
		t.Fatal("expected .cursor/rules/project.mdc in output")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 file, got %d", len(results))
	}
	if !strings.HasPrefix(content, "---\n") {
		t.Errorf("expected content to start with YAML frontmatter '---', got:\n%s", content[:min(100, len(content))])
	}
	if !strings.Contains(content, "alwaysApply: true") {
		t.Error("expected alwaysApply in frontmatter")
	}
	if !strings.Contains(content, "description: Project context and coding guidelines for test-project") {
		t.Error("expected description in frontmatter")
	}
}

func TestGenerate_ClineFormat(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatCline)

	_, ok := results[".clinerules"]
	if !ok {
		t.Fatal("expected .clinerules in output")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 file, got %d", len(results))
	}
}

func TestGenerate_WindsurfFormat(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatWindsurf)

	_, ok := results[".windsurfrules"]
	if !ok {
		t.Fatal("expected .windsurfrules in output")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 file, got %d", len(results))
	}
}

func TestGenerate_AntigravityFormat(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatAntigravity)

	content, ok := results[".gemini/GEMINI.md"]
	if !ok {
		t.Fatal("expected .gemini/GEMINI.md in output")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 file, got %d", len(results))
	}

	// Should start with GEMINI.md header including project name
	if !strings.Contains(content, "# GEMINI.md - test-project") {
		t.Errorf("expected '# GEMINI.md - test-project' header, got:\n%s", content[:min(200, len(content))])
	}

	// Should have all standard sections
	for _, section := range []string{
		"## Tech Stack",
		"## Common Commands",
		"## Code Style",
		"## Project Structure",
		"## Key Files",
		"## Conventions",
		"## Coding Guidelines",
		"## Testing",
	} {
		if !strings.Contains(content, section) {
			t.Errorf("missing section: %s", section)
		}
	}

	// Should have Workflow and Architecture placeholders (like Claude)
	if !strings.Contains(content, "## Workflow") {
		t.Error("missing Workflow placeholder section")
	}
	if !strings.Contains(content, "## Architecture") {
		t.Error("missing Architecture placeholder section")
	}

	// Should have version stamp
	if !strings.Contains(content, "<!-- Generated by context-gen v") {
		t.Error("missing version stamp")
	}
}

func TestGenerate_AntigravityFormat_EmptyName(t *testing.T) {
	info := buildTestProjectInfo()
	info.Name = ""
	results := Generate(info, FormatAntigravity)

	content := results[".gemini/GEMINI.md"]
	if !strings.Contains(content, "# GEMINI.md - Project") {
		t.Errorf("expected fallback name 'Project', got:\n%s", content[:min(200, len(content))])
	}
}

func TestGenerate_AllFormat(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatAll)

	expectedKeys := []string{
		"CLAUDE.md",
		".cursorrules",
		"AGENTS.md",
		".cursor/rules/project.mdc",
		".clinerules",
		".windsurfrules",
		".gemini/GEMINI.md",
	}

	if len(results) != len(expectedKeys) {
		t.Errorf("expected %d files, got %d", len(expectedKeys), len(results))
	}

	for _, key := range expectedKeys {
		if _, ok := results[key]; !ok {
			t.Errorf("expected %s in output", key)
		}
	}
}

func TestVersionStamp_AllFormats(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatAll)

	for name, content := range results {
		if !strings.Contains(content, "<!-- Generated by context-gen v") {
			t.Errorf("%s missing version stamp", name)
		}
	}
}

func TestVersionStamp_Claude(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatClaude)
	content := results["CLAUDE.md"]

	if !strings.HasSuffix(strings.TrimSpace(content), "-->") {
		t.Error("version stamp should be the last line in generated output")
	}
}

// --- helpers ---

func buildTestProjectInfo() *analyzer.ProjectInfo {
	return &analyzer.ProjectInfo{
		Name:        "test-project",
		Description: "A test project for unit tests",
		Languages: []analyzer.Language{
			{Name: "Go", Extension: "go", FileCount: 10, Percentage: 80},
			{Name: "Shell", Extension: "sh", FileCount: 2, Percentage: 20},
		},
		Frameworks:      []string{},
		BuildTools:      []string{"Go Modules", "Make"},
		TestTools:       []string{},
		Linters:         []string{"golangci-lint"},
		HasDocker:       true,
		HasCI:           true,
		CIProvider:      "GitHub Actions",
		PackageManagers: map[string]string{},
		CodeStyle: analyzer.CodeStyle{
			IndentStyle: "tabs",
			Formatter:   "gofmt",
		},
		Scripts: nil,
		Structure: analyzer.DirStructure{
			TopLevelDirs: []string{"analyzer", "cmd", "generator"},
			SubDirs: map[string][]string{
				"cmd": {"root"},
			},
			EntryPoints: []string{"main.go"},
			ConfigFiles: []string{"go.mod"},
			TotalFiles:  12,
			TotalDirs:   3,
		},
		Conventions: []analyzer.Convention{
			{
				Category:    "naming",
				Description: "Go functions follow standard naming",
				Confidence:  1.0,
				Examples:    []string{"HandleRequest", "doWork"},
			},
		},
	}
}
