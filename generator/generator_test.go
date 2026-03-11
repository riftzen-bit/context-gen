package generator

import (
	"strings"
	"testing"

	"github.com/paul/context-gen/analyzer"
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

	// Check expected sections.
	sections := []string{
		"# CLAUDE.md",
		"## Project Overview",
		"## Tech Stack",
		"## Common Commands",
		"## Project Structure",
		"## Conventions",
	}
	for _, section := range sections {
		if !strings.Contains(content, section) {
			t.Errorf("CLAUDE.md missing section: %s", section)
		}
	}

	// Check language info.
	if !strings.Contains(content, "Go") {
		t.Error("expected Go language in output")
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

	// Cursor uses "# Project Context" instead of "# CLAUDE.md".
	if !strings.Contains(content, "# Project Context") {
		t.Error(".cursorrules missing '# Project Context' header")
	}
	if strings.Contains(content, "# CLAUDE.md") {
		t.Error(".cursorrules should not contain '# CLAUDE.md'")
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
		name     string
		pm       string
		wantPM   string
		wantRun  string
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

func TestWriteBuildCommands_PythonPdm(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Python", FileCount: 5, Percentage: 100},
		},
		PackageManagers: map[string]string{"python": "pdm"},
	}

	var b strings.Builder
	writeBuildCommands(&b, info)
	output := b.String()

	if !strings.Contains(output, "pdm install") {
		t.Error("missing pdm install command")
	}
	if !strings.Contains(output, "pdm run pytest") {
		t.Error("missing pdm run pytest command")
	}
}

func TestWriteBuildCommands_PythonPip(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Python", FileCount: 5, Percentage: 100},
		},
		PackageManagers: map[string]string{"python": "pip"},
	}

	var b strings.Builder
	writeBuildCommands(&b, info)
	output := b.String()

	if !strings.Contains(output, "pip install -r requirements.txt") {
		t.Error("missing pip install command")
	}
	if !strings.Contains(output, "pytest") {
		t.Error("missing pytest command")
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

func TestWriteBuildCommands_MultiLanguage(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Go", FileCount: 10, Percentage: 50},
			{Name: "TypeScript", FileCount: 10, Percentage: 50},
		},
		PackageManagers: map[string]string{"js": "pnpm"},
	}

	var b strings.Builder
	writeBuildCommands(&b, info)
	output := b.String()

	// Should have both Go and Node commands.
	if !strings.Contains(output, "go build") {
		t.Error("missing Go build command")
	}
	if !strings.Contains(output, "pnpm install") {
		t.Error("missing pnpm install command")
	}
}

func TestWriteBuildCommands_NoJSCommandsWithoutJSLang(t *testing.T) {
	// If there's a JS package manager but no JS/TS language, no JS commands.
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Go", FileCount: 10, Percentage: 100},
		},
		PackageManagers: map[string]string{"js": "npm"},
	}

	var b strings.Builder
	writeBuildCommands(&b, info)
	output := b.String()

	if strings.Contains(output, "npm install") {
		t.Error("should not have npm commands without JS language")
	}
}

func TestWriteProjectOverview(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Go", FileCount: 10, Percentage: 80},
			{Name: "Shell", FileCount: 2, Percentage: 20},
		},
		Frameworks: []string{"Express", "React"},
		Structure: analyzer.DirStructure{
			TotalFiles: 12,
			TotalDirs:  3,
		},
		PackageManagers: map[string]string{},
	}

	var b strings.Builder
	writeProjectOverview(&b, info)
	output := b.String()

	if !strings.Contains(output, "Go") {
		t.Error("expected primary language Go")
	}
	if !strings.Contains(output, "80%") {
		t.Error("expected 80% percentage")
	}
	if !strings.Contains(output, "Express, React") {
		t.Error("expected frameworks list")
	}
	if !strings.Contains(output, "Files: 12") {
		t.Error("expected file count")
	}
}

func TestWriteTechStack(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Go", FileCount: 10, Percentage: 100},
		},
		BuildTools:      []string{"Go Modules", "Make"},
		TestTools:        []string{"testing"},
		Linters:          []string{"golangci-lint"},
		HasDocker:        true,
		HasCI:            true,
		CIProvider:       "GitHub Actions",
		PackageManagers:  map[string]string{"go": "go modules"},
	}

	var b strings.Builder
	writeTechStack(&b, info)
	output := b.String()

	if !strings.Contains(output, "### Languages") {
		t.Error("missing Languages section")
	}
	if !strings.Contains(output, "### Build Tools") {
		t.Error("missing Build Tools section")
	}
	if !strings.Contains(output, "### Testing") {
		t.Error("missing Testing section")
	}
	if !strings.Contains(output, "### Linting / Formatting") {
		t.Error("missing Linting section")
	}
	if !strings.Contains(output, "Docker") {
		t.Error("missing Docker")
	}
	if !strings.Contains(output, "GitHub Actions") {
		t.Error("missing CI provider")
	}
}

func TestWriteStructure(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Structure: analyzer.DirStructure{
			TopLevelDirs: []string{"cmd", "internal", "pkg"},
			EntryPoints:  []string{"cmd/main.go"},
		},
		PackageManagers: map[string]string{},
	}

	var b strings.Builder
	writeStructure(&b, info)
	output := b.String()

	if !strings.Contains(output, "## Project Structure") {
		t.Error("missing Project Structure section")
	}
	if !strings.Contains(output, "cmd/") {
		t.Error("missing cmd/ directory")
	}
	if !strings.Contains(output, "internal/") {
		t.Error("missing internal/ directory")
	}
	if !strings.Contains(output, "### Entry Points") {
		t.Error("missing Entry Points section")
	}
	if !strings.Contains(output, "cmd/main.go") {
		t.Error("missing main.go entry point")
	}
}

func TestWriteStructure_EmptyDirs(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Structure:       analyzer.DirStructure{},
		PackageManagers: map[string]string{},
	}

	var b strings.Builder
	writeStructure(&b, info)
	output := b.String()

	// Should not write structure section if no top-level dirs.
	if strings.Contains(output, "## Project Structure") {
		t.Error("should not write structure section with no dirs")
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
	if !strings.Contains(output, "my-component.ts") {
		t.Error("missing example")
	}
}

func TestWriteConventions_Empty(t *testing.T) {
	info := &analyzer.ProjectInfo{
		PackageManagers: map[string]string{},
	}

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
		name   string
		langs  []analyzer.Language
		want   bool
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

// --- helpers ---

func buildTestProjectInfo() *analyzer.ProjectInfo {
	return &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Go", Extension: "go", FileCount: 10, Percentage: 80},
			{Name: "Shell", Extension: "sh", FileCount: 2, Percentage: 20},
		},
		Frameworks:  []string{},
		BuildTools:  []string{"Go Modules", "Make"},
		TestTools:   []string{},
		Linters:     []string{},
		HasDocker:   true,
		HasCI:       true,
		CIProvider:  "GitHub Actions",
		PackageManagers: map[string]string{},
		Structure: analyzer.DirStructure{
			TopLevelDirs: []string{"analyzer", "generator"},
			EntryPoints:  []string{"main.go"},
			ConfigFiles:  []string{"go.mod"},
			TotalFiles:   12,
			TotalDirs:    2,
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
