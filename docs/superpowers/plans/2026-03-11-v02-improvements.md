# context-gen v0.2 Improvements Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade context-gen with Anthropic-compliant CLAUDE.md output, modern charmbracelet CLI UI, and MIT license.

**Architecture:** Three sequential phases. Phase 1 modifies the data layer (analyzer) and template layer (generator). Phase 2 replaces the presentation layer (cmd/interactive, new ui/ package). Phase 3 adds license and bumps version. Each phase produces a working, testable binary.

**Tech Stack:** Go 1.26, charmbracelet/lipgloss, charmbracelet/bubbletea, charmbracelet/bubbles

**Spec:** `docs/superpowers/specs/2026-03-11-v02-improvements-design.md`

---

## Chunk 1: Phase 1 — CLAUDE.md Anthropic Compliance

### File Map

| Action | File | Responsibility |
|--------|------|----------------|
| Modify | `analyzer/types.go` | Add Name, Description, CodeStyle, Scripts fields |
| Create | `analyzer/config_reader.go` | Parse .editorconfig, .prettierrc, README.md, Makefile, package.json scripts |
| Create | `analyzer/config_reader_test.go` | Unit tests for config parsing |
| Modify | `analyzer/detector.go` | Call config reader, detect project name & formatter |
| Modify | `analyzer/detector_test.go` | Update tests for new fields |
| Modify | `analyzer/scanner.go` | Add new config files to scan list |
| Modify | `generator/generator.go` | Rewrite template to Anthropic format |
| Modify | `generator/generator_test.go` | Update tests for new template |
| Create | `testdata/` | Fixture files for config reader tests |

---

### Task 1: Add new types to analyzer/types.go

**Files:**
- Modify: `analyzer/types.go`

- [ ] **Step 1: Add CodeStyle struct and new ProjectInfo fields**

```go
// In analyzer/types.go, add after Convention struct:

// CodeStyle holds detected code formatting rules.
type CodeStyle struct {
	IndentStyle string // "tabs" or "spaces"
	IndentSize  int    // 2, 4, etc.
	LineLength  int    // 80, 100, 120, etc.
	Formatter   string // "gofmt", "prettier", "black", "rustfmt", etc.
	ExtraRules  []string
}

// Add to ProjectInfo struct:
//   Name        string
//   Description string
//   CodeStyle   CodeStyle
//   Scripts     map[string]string
```

Actual edit: add these four fields to the existing `ProjectInfo` struct after `Conventions`.

- [ ] **Step 2: Verify build**

Run: `cd /home/paul/Projects/context-gen && go build ./...`
Expected: BUILD SUCCESS (new fields are zero-valued by default, no existing code breaks)

- [ ] **Step 3: Commit**

```bash
git add analyzer/types.go
git commit -m "feat: add Name, Description, CodeStyle, Scripts fields to ProjectInfo"
```

---

### Task 2: Register new config files in scanner

**Files:**
- Modify: `analyzer/scanner.go`

- [ ] **Step 1: Add missing config files to configFiles map**

Add these entries to the `configFiles` map in `scanner.go`:

```go
".editorconfig":    true,
"README.md":        true,
"readme.md":        true,
"README":           true,
"rustfmt.toml":     true,
```

- [ ] **Step 2: Verify build and existing tests pass**

Run: `cd /home/paul/Projects/context-gen && go test ./analyzer/ -v -run TestScan`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add analyzer/scanner.go
git commit -m "feat: register .editorconfig, README.md, rustfmt.toml in scanner"
```

---

### Task 3: Create config reader with tests (TDD)

**Files:**
- Create: `analyzer/config_reader.go`
- Create: `analyzer/config_reader_test.go`
- Create: `analyzer/testdata/editorconfig_sample` (fixture)
- Create: `analyzer/testdata/prettierrc_sample.json` (fixture)
- Create: `analyzer/testdata/readme_sample.md` (fixture)
- Create: `analyzer/testdata/makefile_sample` (fixture)
- Create: `analyzer/testdata/package_sample.json` (fixture)
- Create: `analyzer/testdata/gomod_sample` (fixture)

- [ ] **Step 1: Create test fixtures**

Create `analyzer/testdata/` directory with these fixtures:

`analyzer/testdata/editorconfig_sample`:
```ini
root = true

[*]
indent_style = tab
indent_size = 4
max_line_length = 120
```

`analyzer/testdata/prettierrc_sample.json`:
```json
{
  "printWidth": 100,
  "tabWidth": 2,
  "singleQuote": true,
  "semi": false
}
```

`analyzer/testdata/readme_sample.md`:
```markdown
# My Project

[![Build](https://badge.example.com)]

A CLI tool that does amazing things for developers.

## Installation

Run `go install`.
```

`analyzer/testdata/makefile_sample`:
```makefile
.PHONY: build test lint

build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run

dev:
	go run .
```

`analyzer/testdata/package_sample.json`:
```json
{
  "name": "my-app",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "test": "vitest",
    "lint": "eslint ."
  }
}
```

`analyzer/testdata/gomod_sample`:
```
module github.com/test/myproject

go 1.21
```

- [ ] **Step 2: Write failing tests for config reader**

Create `analyzer/config_reader_test.go`:

```go
package analyzer

import (
	"path/filepath"
	"runtime"
	"testing"
)

func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata")
}

func TestReadEditorConfig(t *testing.T) {
	style := readEditorConfig(filepath.Join(testdataDir(), "editorconfig_sample"))

	if style.IndentStyle != "tab" {
		t.Errorf("IndentStyle = %q, want %q", style.IndentStyle, "tab")
	}
	if style.IndentSize != 4 {
		t.Errorf("IndentSize = %d, want %d", style.IndentSize, 4)
	}
	if style.LineLength != 120 {
		t.Errorf("LineLength = %d, want %d", style.LineLength, 120)
	}
}

func TestReadEditorConfig_NotFound(t *testing.T) {
	style := readEditorConfig("/nonexistent/.editorconfig")
	if style.IndentStyle != "" {
		t.Error("expected empty CodeStyle for missing file")
	}
}

func TestReadPrettierConfig(t *testing.T) {
	rules := readPrettierConfig(filepath.Join(testdataDir(), "prettierrc_sample.json"))

	if len(rules) == 0 {
		t.Fatal("expected rules from prettier config")
	}

	found := false
	for _, r := range rules {
		if r == "printWidth: 100" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'printWidth: 100' in rules, got %v", rules)
	}
}

func TestReadREADME(t *testing.T) {
	desc := readREADMEDescription(filepath.Join(testdataDir(), "readme_sample.md"))

	if desc != "A CLI tool that does amazing things for developers." {
		t.Errorf("description = %q", desc)
	}
}

func TestReadREADME_NotFound(t *testing.T) {
	desc := readREADMEDescription("/nonexistent/README.md")
	if desc != "" {
		t.Error("expected empty description for missing file")
	}
}

func TestReadMakefileTargets(t *testing.T) {
	targets := readMakefileTargets(filepath.Join(testdataDir(), "makefile_sample"))

	want := map[string]bool{"build": true, "test": true, "lint": true, "dev": true}
	for _, target := range targets {
		if !want[target] {
			t.Errorf("unexpected target: %s", target)
		}
	}
	if len(targets) != 4 {
		t.Errorf("got %d targets, want 4", len(targets))
	}
}

func TestReadPackageJSONScripts(t *testing.T) {
	scripts := readPackageJSONScripts(filepath.Join(testdataDir(), "package_sample.json"))

	if scripts["dev"] != "vite" {
		t.Errorf("scripts[dev] = %q, want %q", scripts["dev"], "vite")
	}
	if scripts["build"] != "vite build" {
		t.Errorf("scripts[build] = %q, want %q", scripts["build"], "vite build")
	}
	if scripts["test"] != "vitest" {
		t.Errorf("scripts[test] = %q, want %q", scripts["test"], "vitest")
	}
}

func TestReadPackageJSONName(t *testing.T) {
	name := readPackageJSONName(filepath.Join(testdataDir(), "package_sample.json"))
	if name != "my-app" {
		t.Errorf("name = %q, want %q", name, "my-app")
	}
}

func TestReadGoModName(t *testing.T) {
	name := readGoModName(filepath.Join(testdataDir(), "gomod_sample"))
	if name != "myproject" {
		t.Errorf("name = %q, want %q", name, "myproject")
	}
}

func TestReadGoModName_NotFound(t *testing.T) {
	name := readGoModName("/nonexistent/go.mod")
	if name != "" {
		t.Error("expected empty name for missing file")
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `cd /home/paul/Projects/context-gen && go test ./analyzer/ -v -run "TestRead"`
Expected: FAIL — functions don't exist yet

- [ ] **Step 4: Implement config_reader.go**

Create `analyzer/config_reader.go`:

```go
package analyzer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// readEditorConfig parses a .editorconfig file and returns CodeStyle from the [*] section.
func readEditorConfig(path string) CodeStyle {
	f, err := os.Open(path)
	if err != nil {
		return CodeStyle{}
	}
	defer f.Close()

	var style CodeStyle
	inGlobal := false
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") {
			inGlobal = line == "[*]"
			continue
		}

		if !inGlobal {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "indent_style":
			style.IndentStyle = val
		case "indent_size":
			if n, err := strconv.Atoi(val); err == nil {
				style.IndentSize = n
			}
		case "max_line_length":
			if n, err := strconv.Atoi(val); err == nil {
				style.LineLength = n
			}
		}
	}

	return style
}

// readPrettierConfig parses a .prettierrc or .prettierrc.json and returns style rules as strings.
func readPrettierConfig(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}

	var rules []string
	for _, key := range []string{"printWidth", "tabWidth", "singleQuote", "semi", "trailingComma"} {
		if val, ok := cfg[key]; ok {
			rules = append(rules, fmt.Sprintf("%s: %v", key, val))
		}
	}

	return rules
}

// readREADMEDescription extracts the first descriptive paragraph from a README.md.
// Skips headings (#), badges ([![), links ([), images (!), HTML (<), and empty lines.
func readREADMEDescription(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[![") || strings.HasPrefix(line, "[") {
			continue
		}
		if strings.HasPrefix(line, "!") || strings.HasPrefix(line, "<") {
			continue
		}

		// Found first real text line — collect until empty line or heading.
		var desc strings.Builder
		desc.WriteString(line)
		for scanner.Scan() {
			next := strings.TrimSpace(scanner.Text())
			if next == "" || strings.HasPrefix(next, "#") {
				break
			}
			desc.WriteString(" ")
			desc.WriteString(next)
		}

		result := desc.String()
		if len(result) > 200 {
			result = result[:200]
		}
		return result
	}

	return ""
}

var makeTargetRegex = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*)\s*:`)

// readMakefileTargets extracts target names from a Makefile.
func readMakefileTargets(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var targets []string
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		matches := makeTargetRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			target := matches[1]
			if !seen[target] {
				seen[target] = true
				targets = append(targets, target)
			}
		}
	}

	return targets
}

// readPackageJSONScripts reads the "scripts" field from a package.json.
func readPackageJSONScripts(path string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	return pkg.Scripts
}

// readPackageJSONName reads the "name" field from a package.json.
func readPackageJSONName(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	var pkg struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return ""
	}

	return pkg.Name
}

// readGoModName reads the module name from go.mod and returns the last path segment.
func readGoModName(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			mod := strings.TrimPrefix(line, "module ")
			mod = strings.TrimSpace(mod)
			parts := strings.Split(mod, "/")
			return parts[len(parts)-1]
		}
	}

	return ""
}

// readCargoTomlName reads the package name from Cargo.toml using simple string parsing.
func readCargoTomlName(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[1])
				name = strings.Trim(name, `"'`)
				return name
			}
		}
	}
	return ""
}

// readPyprojectTomlName reads the project name from pyproject.toml.
func readPyprojectTomlName(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	inProject := false
	for _, line := range strings.Split(string(data), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			inProject = trimmed == "[project]"
			continue
		}
		if inProject && strings.HasPrefix(trimmed, "name") {
			parts := strings.SplitN(trimmed, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[1])
				name = strings.Trim(name, `"'`)
				return name
			}
		}
	}
	return ""
}

// hasPyprojectBlack checks if pyproject.toml has a [tool.black] section.
func hasPyprojectBlack(root string, scan *ScanResult) bool {
	for _, f := range scan.Files {
		if f.Name == "pyproject.toml" {
			data, err := os.ReadFile(filepath.Join(root, f.Path))
			if err != nil {
				return false
			}
			return strings.Contains(string(data), "[tool.black]")
		}
	}
	return false
}

// detectProjectName determines the project name from available config files.
// Priority: package.json name > go.mod module > Cargo.toml name > pyproject.toml name > directory name.
func detectProjectName(root string, scan *ScanResult) string {
	for _, f := range scan.Files {
		if f.Name == "package.json" && f.Path == "package.json" {
			if name := readPackageJSONName(filepath.Join(root, f.Path)); name != "" {
				return name
			}
		}
	}

	for _, f := range scan.Files {
		if f.Name == "go.mod" {
			if name := readGoModName(filepath.Join(root, f.Path)); name != "" {
				return name
			}
		}
	}

	for _, f := range scan.Files {
		if f.Name == "Cargo.toml" {
			if name := readCargoTomlName(filepath.Join(root, f.Path)); name != "" {
				return name
			}
		}
	}

	for _, f := range scan.Files {
		if f.Name == "pyproject.toml" {
			if name := readPyprojectTomlName(filepath.Join(root, f.Path)); name != "" {
				return name
			}
		}
	}

	// Fallback: directory name
	return filepath.Base(root)
}

// detectFormatter returns the default formatter for the primary language.
func detectFormatter(root string, primaryLang string, scan *ScanResult) string {
	// Check for formatter config files first
	for _, f := range scan.Files {
		switch f.Name {
		case ".prettierrc", ".prettierrc.json":
			return "prettier"
		case "rustfmt.toml":
			return "rustfmt"
		}
	}

	// Check pyproject.toml for black
	if primaryLang == "Python" && hasPyprojectBlack(root, scan) {
		return "black"
	}

	switch primaryLang {
	case "Go":
		return "gofmt"
	case "Rust":
		return "rustfmt"
	default:
		return ""
	}
}

// detectDescription reads project description from README.md.
func detectDescription(root string, scan *ScanResult) string {
	for _, f := range scan.Files {
		switch f.Name {
		case "README.md", "readme.md", "README":
			if desc := readREADMEDescription(filepath.Join(root, f.Path)); desc != "" {
				return desc
			}
		}
	}
	return ""
}

// detectCodeStyle reads .editorconfig and prettier config for code style rules.
func detectCodeStyle(root string, scan *ScanResult, primaryLang string) CodeStyle {
	var style CodeStyle

	// Try .editorconfig first
	for _, f := range scan.Files {
		if f.Name == ".editorconfig" {
			style = readEditorConfig(filepath.Join(root, f.Path))
			break
		}
	}

	style.Formatter = detectFormatter(root, primaryLang, scan)

	// Read prettier for extra rules
	for _, f := range scan.Files {
		if f.Name == ".prettierrc" || f.Name == ".prettierrc.json" {
			style.ExtraRules = readPrettierConfig(filepath.Join(root, f.Path))
			break
		}
	}

	return style
}

// detectScripts reads build/test/lint commands from package.json scripts or Makefile.
// Returns nil if no scripts found (fallback to hardcoded commands).
func detectScripts(root string, scan *ScanResult) map[string]string {
	// Try package.json scripts first
	for _, f := range scan.Files {
		if f.Name == "package.json" && f.Path == "package.json" {
			scripts := readPackageJSONScripts(filepath.Join(root, f.Path))
			if len(scripts) > 0 {
				return scripts
			}
		}
	}

	// Try Makefile targets
	for _, f := range scan.Files {
		if f.Name == "Makefile" && f.Path == "Makefile" {
			targets := readMakefileTargets(filepath.Join(root, f.Path))
			if len(targets) > 0 {
				scripts := make(map[string]string)
				for _, t := range targets {
					scripts[t] = fmt.Sprintf("make %s", t)
				}
				return scripts
			}
		}
	}

	return nil
}
```

**Note on tsconfig.json and .eslintrc.json:** These are listed in the spec's config table but are deferred from this implementation. The value/effort ratio is low — tsconfig strict mode and ESLint rule overrides rarely produce actionable CLAUDE.md content. The `.editorconfig` and `.prettierrc` parsers cover the most impactful code style rules. tsconfig/eslint parsing can be added in a future iteration if needed.

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /home/paul/Projects/context-gen && go test ./analyzer/ -v -run "TestRead"`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add analyzer/config_reader.go analyzer/config_reader_test.go testdata/
git commit -m "feat: add config reader for .editorconfig, prettier, README, Makefile, package.json"
```

---

### Task 4: Integrate config reader into detector

**Files:**
- Modify: `analyzer/detector.go`

- [ ] **Step 1: Write test for new ProjectInfo fields**

Add to `analyzer/detector_test.go`:

```go
func TestDetect_PopulatesName(t *testing.T) {
	// Create a temp dir with a go.mod
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/test/myproject\n\ngo 1.21\n"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)

	scan, _ := Scan(dir)
	info, err := Detect(dir, scan)
	if err != nil {
		t.Fatal(err)
	}

	if info.Name != "myproject" {
		t.Errorf("Name = %q, want %q", info.Name, "myproject")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /home/paul/Projects/context-gen && go test ./analyzer/ -v -run TestDetect_PopulatesName`
Expected: FAIL

- [ ] **Step 3: Add config reader calls to Detect function**

In `analyzer/detector.go`, update the `Detect` function. After `detectDocker(scan, info)` and before conventions, add:

```go
	// Detect project metadata from config files
	info.Name = detectProjectName(root, scan)
	info.Description = detectDescription(root, scan)

	primaryLang := ""
	if len(info.Languages) > 0 {
		primaryLang = info.Languages[0].Name
	}
	info.CodeStyle = detectCodeStyle(root, scan, primaryLang)
	info.Scripts = detectScripts(root, scan)
```

Also initialize `Scripts` in the info creation: `Scripts: make(map[string]string),` — but only as fallback if detectScripts returns nil.

- [ ] **Step 4: Run all tests**

Run: `cd /home/paul/Projects/context-gen && go test ./analyzer/ -v`
Expected: ALL PASS

- [ ] **Step 5: Commit**

```bash
git add analyzer/detector.go analyzer/detector_test.go
git commit -m "feat: integrate config reader into detection pipeline"
```

---

### Task 5: Rewrite generator template (TDD)

**Files:**
- Modify: `generator/generator.go`
- Modify: `generator/generator_test.go`

- [ ] **Step 1: Write new tests for Anthropic-compliant output**

Replace test expectations in `generator/generator_test.go`. Key changes:

```go
func TestGenerate_ClaudeFormat(t *testing.T) {
	info := buildTestProjectInfo()
	results := Generate(info, FormatClaude)

	content, ok := results["CLAUDE.md"]
	if !ok {
		t.Fatal("expected CLAUDE.md in output")
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

	// Should NOT have the old verbose sections
	if strings.Contains(content, "### Languages") {
		t.Error("should not have verbose Languages subsection")
	}
	if strings.Contains(content, "### Build Tools") {
		t.Error("should not have verbose Build Tools subsection")
	}
}
```

Update `buildTestProjectInfo()`:
```go
func buildTestProjectInfo() *analyzer.ProjectInfo {
	return &analyzer.ProjectInfo{
		Name: "test-project",
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
			EntryPoints:  []string{"main.go"},
			ConfigFiles:  []string{"go.mod"},
			TotalFiles:   12,
			TotalDirs:    3,
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
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /home/paul/Projects/context-gen && go test ./generator/ -v -run TestGenerate_ClaudeFormat`
Expected: FAIL

- [ ] **Step 3: Rewrite generator.go template functions**

Replace `generateClaude`, `generateCursor`, and all write functions in `generator/generator.go`:

```go
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
		for name, cmd := range info.Scripts {
			b.WriteString(fmt.Sprintf("%-25s # %s\n", cmd, name))
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
```

Keep the existing `writeBuildCommands`, `hasLang`, `hasJSLang` functions as fallback (they're used when no scripts are detected).

Remove the old `writeProjectOverview` and `writeTechStack` functions (replaced by new ones).
Remove the old `writeStructure` function (replaced by `writeStructureAnnotated`).

- [ ] **Step 4: Update remaining tests**

Update all tests in `generator_test.go` to match new template format:

**DELETE these tests** (they call functions that no longer exist):
- `TestWriteProjectOverview` — calls deleted `writeProjectOverview()`
- `TestWriteTechStack` — calls deleted `writeTechStack()`
- `TestWriteStructure` — calls deleted `writeStructure()`
- `TestWriteStructure_EmptyDirs` — calls deleted `writeStructure()`

**REPLACE with new tests:**
- `TestWriteHeader` — test new `writeHeader()` function, verify project name in Claude format, "# Project Context" in Cursor format
- `TestWriteTechStackConcise` — test pipe-separated output format
- `TestWriteStructureAnnotated` — test directory annotations (e.g. `cmd/` → "CLI entry points"), test that unknown dirs get no annotation
- `TestWriteStructureAnnotated_EmptyDirs` — verify no output when no dirs

**UPDATE these tests:**
- `TestGenerate_CursorFormat`: check for `# Project Context`, verify NO `## Workflow` or `## Architecture` sections
- `TestWriteBuildCommands_*` — keep as-is, they test the fallback path which is still valid
- **ADD** `TestWriteCommands_WithScripts` — test that when `info.Scripts` is non-nil, detected scripts are used instead of hardcoded commands

**FIX map iteration order in `writeCommands`:** Sort script keys before iterating to ensure deterministic output:
```go
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
```
Add `"sort"` to generator.go imports.

- [ ] **Step 5: Run all tests**

Run: `cd /home/paul/Projects/context-gen && go test ./... -v`
Expected: ALL PASS

- [ ] **Step 6: Verify output is under 200 lines**

Run: `cd /home/paul/Projects/context-gen && go run . preview -d . -f claude 2>&1 | wc -l`
Expected: < 200

- [ ] **Step 7: Commit**

```bash
git add generator/generator.go generator/generator_test.go
git commit -m "feat: rewrite generator template to follow Anthropic CLAUDE.md guidelines"
```

---

## Chunk 2: Phase 2 — Charmbracelet CLI UI

### Task 6: Add charmbracelet dependencies

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Add dependencies**

```bash
cd /home/paul/Projects/context-gen
go get github.com/charmbracelet/lipgloss@latest
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
```

- [ ] **Step 2: Verify build**

Run: `go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add charmbracelet dependencies (lipgloss, bubbletea, bubbles)"
```

---

### Task 7: Create ui/styles.go

**Files:**
- Create: `ui/styles.go`

- [ ] **Step 1: Create the styles file**

```go
package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorPrimary   = lipgloss.Color("39")  // blue
	ColorSuccess   = lipgloss.Color("82")  // green
	ColorWarning   = lipgloss.Color("214") // yellow
	ColorError     = lipgloss.Color("196") // red
	ColorSubtle    = lipgloss.Color("241") // gray
	ColorHighlight = lipgloss.Color("212") // pink

	// Text styles
	Bold      = lipgloss.NewStyle().Bold(true)
	Subtle    = lipgloss.NewStyle().Foreground(ColorSubtle)
	Success   = lipgloss.NewStyle().Foreground(ColorSuccess)
	Warning   = lipgloss.NewStyle().Foreground(ColorWarning)
	Error     = lipgloss.NewStyle().Foreground(ColorError)
	Highlight = lipgloss.NewStyle().Foreground(ColorHighlight)

	// Banner box
	BannerStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 2)

	// Results box
	ResultsStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(0, 1)

	// Menu item styles
	MenuSelected   = lipgloss.NewStyle().Foreground(ColorHighlight).Bold(true)
	MenuUnselected = lipgloss.NewStyle().Foreground(ColorSubtle)

	// Symbols
	SymbolSuccess = Success.Render("✓")
	SymbolError   = Error.Render("✗")
	SymbolArrow   = Highlight.Render(">")
)

// Version is the application version string.
const Version = "0.2.0"

// Banner returns the styled welcome banner.
func Banner() string {
	content := Bold.Render("context-gen") + " v" + Version + "\n" +
		Subtle.Render("Generate AI context files")
	return BannerStyle.Render(content)
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 3: Commit**

```bash
git add ui/styles.go
git commit -m "feat: add lipgloss UI styles and banner"
```

---

### Task 8: Create bubbletea menu model

**Files:**
- Create: `ui/menu.go`
- Create: `ui/menu_test.go`

- [ ] **Step 1: Write failing test**

Create `ui/menu_test.go`:

```go
package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMenuModel_Navigation(t *testing.T) {
	m := NewMenuModel()

	// Initial cursor at 0
	if m.cursor != 0 {
		t.Errorf("initial cursor = %d, want 0", m.cursor)
	}

	// Move down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(MenuModel)
	if m.cursor != 1 {
		t.Errorf("after down, cursor = %d, want 1", m.cursor)
	}

	// Move up
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(MenuModel)
	if m.cursor != 0 {
		t.Errorf("after up, cursor = %d, want 0", m.cursor)
	}

	// Wrap around at top
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(MenuModel)
	if m.cursor != len(m.items)-1 {
		t.Errorf("wrap top, cursor = %d, want %d", m.cursor, len(m.items)-1)
	}
}

func TestMenuModel_Select(t *testing.T) {
	m := NewMenuModel()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = updated.(MenuModel)

	if m.choice != 0 {
		t.Errorf("choice = %d, want 0", m.choice)
	}
	// Should return quit command (menu exits after selection)
	if cmd == nil {
		t.Error("expected quit command after enter")
	}
}
```

- [ ] **Step 2: Implement menu model**

Create `ui/menu.go`:

```go
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// MenuChoice represents the selected menu action.
type MenuChoice int

const (
	ChoiceGenerate MenuChoice = iota
	ChoiceUpdate
	ChoicePreview
	ChoiceHelp
	ChoiceExit
	ChoiceNone MenuChoice = -1
)

// MenuModel is the bubbletea model for the main menu.
type MenuModel struct {
	items  []string
	cursor int
	choice MenuChoice
	done   bool
}

// NewMenuModel creates a new menu with default items.
func NewMenuModel() MenuModel {
	return MenuModel{
		items: []string{
			"Generate context files",
			"Update existing files",
			"Preview output",
			"Help",
			"Exit",
		},
		choice: ChoiceNone,
	}
}

func (m MenuModel) Init() tea.Cmd { return nil }

func (m MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp, tea.KeyShiftTab:
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.items) - 1
			}
		case tea.KeyDown, tea.KeyTab:
			m.cursor++
			if m.cursor >= len(m.items) {
				m.cursor = 0
			}
		case tea.KeyEnter:
			m.choice = MenuChoice(m.cursor)
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.choice = ChoiceExit
			m.done = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m MenuModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(Bold.Render("What would you like to do?"))
	b.WriteString("\n\n")

	for i, item := range m.items {
		cursor := "  "
		style := MenuUnselected
		if i == m.cursor {
			cursor = SymbolArrow + " "
			style = MenuSelected
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(item)))
	}

	b.WriteString("\n")
	b.WriteString(Subtle.Render("↑/↓ navigate • enter select • esc quit"))
	b.WriteString("\n")

	return b.String()
}

// Choice returns the selected menu choice.
func (m MenuModel) Choice() MenuChoice { return m.choice }
```

- [ ] **Step 3: Run tests**

Run: `cd /home/paul/Projects/context-gen && go test ./ui/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add ui/menu.go ui/menu_test.go
git commit -m "feat: add bubbletea menu model with arrow key navigation"
```

---

### Task 9: Create spinner and results components

**Files:**
- Create: `ui/spinner.go`
- Create: `ui/results.go`

- [ ] **Step 1: Create spinner model**

Create `ui/spinner.go`:

```go
package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ScanDoneMsg signals that scanning completed.
type ScanDoneMsg struct{ Err error }

// SpinnerModel shows a spinner while scanning.
type SpinnerModel struct {
	spinner spinner.Model
	message string
	done    bool
	err     error
}

// NewSpinnerModel creates a spinner with a message.
func NewSpinnerModel(msg string) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(ColorPrimary)
	return SpinnerModel{spinner: s, message: msg}
}

func (m SpinnerModel) Init() tea.Cmd { return m.spinner.Tick }

func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ScanDoneMsg:
		m.done = true
		m.err = msg.Err
		return m, tea.Quit
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m SpinnerModel) View() string {
	if m.done {
		if m.err != nil {
			return fmt.Sprintf("  %s %s\n", SymbolError, Error.Render(m.err.Error()))
		}
		return fmt.Sprintf("  %s Done\n", SymbolSuccess)
	}
	return fmt.Sprintf("  %s %s\n", m.spinner.View(), m.message)
}
```

- [ ] **Step 2: Create results display**

Create `ui/results.go`:

```go
package ui

import (
	"fmt"
	"strings"

	"github.com/paul/context-gen/analyzer"
)

// FormatResults renders detection results in a styled box.
func FormatResults(info *analyzer.ProjectInfo) string {
	var rows []string

	if len(info.Languages) > 0 {
		var langs []string
		for _, l := range info.Languages {
			langs = append(langs, fmt.Sprintf("%s (%.0f%%)", l.Name, l.Percentage))
		}
		rows = append(rows, fmt.Sprintf("%-12s %s", "Languages", strings.Join(langs, ", ")))
	}

	if len(info.BuildTools) > 0 {
		rows = append(rows, fmt.Sprintf("%-12s %s", "Build", strings.Join(info.BuildTools, ", ")))
	}

	if len(info.TestTools) > 0 {
		rows = append(rows, fmt.Sprintf("%-12s %s", "Testing", strings.Join(info.TestTools, ", ")))
	}

	if len(info.Linters) > 0 {
		rows = append(rows, fmt.Sprintf("%-12s %s", "Linting", strings.Join(info.Linters, ", ")))
	}

	if info.HasDocker {
		rows = append(rows, fmt.Sprintf("%-12s %s", "Docker", "yes"))
	}

	if info.HasCI {
		rows = append(rows, fmt.Sprintf("%-12s %s", "CI/CD", info.CIProvider))
	}

	content := Bold.Render("Detection Results") + "\n" + strings.Join(rows, "\n")
	return ResultsStyle.Render(content)
}

// FormatFileCreated formats a success message for file creation.
func FormatFileCreated(path string) string {
	return fmt.Sprintf("  %s Created %s", SymbolSuccess, path)
}

// FormatFileUpdated formats a success message for file update.
func FormatFileUpdated(path string) string {
	return fmt.Sprintf("  %s Updated %s", SymbolSuccess, path)
}

// FormatFileSkipped formats a skip message.
func FormatFileSkipped(name, reason string) string {
	return fmt.Sprintf("  %s %s %s", Warning.Render("~"), name, Subtle.Render(reason))
}

// FormatError formats an error message.
func FormatError(msg string) string {
	return fmt.Sprintf("  %s %s", SymbolError, Error.Render(msg))
}
```

- [ ] **Step 3: Verify build**

Run: `go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 4: Commit**

```bash
git add ui/spinner.go ui/results.go
git commit -m "feat: add spinner and results display components"
```

---

### Task 10: Create input prompts

**Files:**
- Create: `ui/prompts.go`

- [ ] **Step 1: Create styled prompts**

Create `ui/prompts.go`:

```go
package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// PromptModel is a text input prompt with a default value.
type PromptModel struct {
	label      string
	defaultVal string
	value      string
	done       bool
}

// NewPromptModel creates a new text prompt.
func NewPromptModel(label, defaultVal string) PromptModel {
	return PromptModel{label: label, defaultVal: defaultVal}
}

func (m PromptModel) Init() tea.Cmd { return nil }

func (m PromptModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.value == "" {
				m.value = m.defaultVal
			}
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.value = m.defaultVal
			m.done = true
			return m, tea.Quit
		case tea.KeyBackspace:
			if len(m.value) > 0 {
				m.value = m.value[:len(m.value)-1]
			}
		default:
			if msg.Type == tea.KeyRunes {
				m.value += string(msg.Runes)
			}
		}
	}
	return m, nil
}

func (m PromptModel) View() string {
	var b strings.Builder

	if m.defaultVal != "" {
		b.WriteString(fmt.Sprintf("\n%s [%s]: %s", Bold.Render(m.label), Subtle.Render(m.defaultVal), m.value))
	} else {
		b.WriteString(fmt.Sprintf("\n%s: %s", Bold.Render(m.label), m.value))
	}

	if !m.done {
		b.WriteString("█") // cursor
	}
	b.WriteString("\n")

	return b.String()
}

// Value returns the entered value.
func (m PromptModel) Value() string { return m.value }

// FormatSelectModel is a choice selector (like the format picker).
type FormatSelectModel struct {
	items  []string
	cursor int
	done   bool
}

// NewFormatSelectModel creates a format selector.
func NewFormatSelectModel() FormatSelectModel {
	return FormatSelectModel{
		items: []string{
			"Claude (CLAUDE.md)",
			"Cursor (.cursorrules)",
			"Both",
		},
	}
}

func (m FormatSelectModel) Init() tea.Cmd { return nil }

func (m FormatSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.items) - 1
			}
		case tea.KeyDown:
			m.cursor++
			if m.cursor >= len(m.items) {
				m.cursor = 0
			}
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cursor = 2 // default to "Both"
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m FormatSelectModel) View() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(Bold.Render("Select format:"))
	b.WriteString("\n\n")

	for i, item := range m.items {
		cursor := "  "
		style := MenuUnselected
		if i == m.cursor {
			cursor = SymbolArrow + " "
			style = MenuSelected
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, style.Render(item)))
	}

	return b.String()
}

// Choice returns 0=claude, 1=cursor, 2=both.
func (m FormatSelectModel) Choice() int { return m.cursor }
```

- [ ] **Step 2: Verify build**

Run: `go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 3: Commit**

```bash
git add ui/prompts.go
git commit -m "feat: add styled text input and format selector prompts"
```

---

### Task 11: Integrate UI into cmd/interactive.go and cmd/root.go

**Files:**
- Modify: `cmd/interactive.go`
- Modify: `cmd/root.go`

- [ ] **Step 1: Rewrite cmd/interactive.go to use bubbletea**

Replace the entire file with the bubbletea-based implementation. The new version:
- Uses `tea.NewProgram(ui.NewMenuModel())` for main menu
- Uses `ui.NewPromptModel()` for text input
- Uses `ui.NewFormatSelectModel()` for format selection
- Uses `ui.NewSpinnerModel()` for scanning spinner
- Falls back to old fmt-based flow if `tea.NewProgram().Run()` fails

Key structure:
```go
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/paul/context-gen/generator"
	"github.com/paul/context-gen/ui"
)

func runInteractive() {
	fmt.Println()
	fmt.Println(ui.Banner())

	for {
		menu := ui.NewMenuModel()
		p := tea.NewProgram(menu)
		result, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "UI error: %v\n", err)
			return
		}

		m := result.(ui.MenuModel)
		switch m.Choice() {
		case ui.ChoiceGenerate:
			interactiveInit()
		case ui.ChoiceUpdate:
			interactiveUpdate()
		case ui.ChoicePreview:
			interactivePreview()
		case ui.ChoiceHelp:
			fmt.Println()
			printUsage()
		case ui.ChoiceExit, ui.ChoiceNone:
			fmt.Println()
			fmt.Println(ui.Success.Render("Bye!"))
			return
		}

		// "Press Enter to continue" prompt
		fmt.Println()
		fmt.Print(ui.Subtle.Render("Press Enter to continue..."))
		fmt.Scanln()
	}
}

func promptDir() string {
	p := tea.NewProgram(ui.NewPromptModel("Target directory", "."))
	result, err := p.Run()
	if err != nil {
		return "."
	}
	return result.(ui.PromptModel).Value()
}

func promptFormatUI() generator.Format {
	p := tea.NewProgram(ui.NewFormatSelectModel())
	result, err := p.Run()
	if err != nil {
		return generator.FormatBoth
	}
	m := result.(ui.FormatSelectModel)
	switch m.Choice() {
	case 0:
		return generator.FormatClaude
	case 1:
		return generator.FormatCursor
	default:
		return generator.FormatBoth
	}
}

func interactiveInit() {
	dir := promptDir()
	format := promptFormatUI()
	args := []string{"-d", dir, "-f", string(format)}
	if err := runInit(args); err != nil {
		fmt.Println(ui.FormatError(err.Error()))
	}
}

func interactiveUpdate() {
	dir := promptDir()
	args := []string{"-d", dir}
	if err := runUpdate(args); err != nil {
		fmt.Println(ui.FormatError(err.Error()))
	}
}

func interactivePreview() {
	dir := promptDir()
	format := promptFormatUI()
	args := []string{"-d", dir, "-f", string(format)}
	if err := runPreview(args); err != nil {
		fmt.Println(ui.FormatError(err.Error()))
	}
}
```

- [ ] **Step 2: Update cmd/root.go output to use lipgloss**

Replace ANSI color constants and helper functions with imports from `ui` package:
- Remove `colorReset`, `colorBold`, etc. constants
- Remove `bold()`, `green()`, `yellow()`, `red()` functions
- Use `ui.Bold.Render()`, `ui.Success.Render()`, etc. throughout
- Update version string to `Version` from `ui` package
- Update `printDetectionSummary` to use `ui.FormatResults()`

- [ ] **Step 2.5: Add TTY detection to ui/styles.go**

Add TTY detection function. When stdout is not a TTY (piped), disable bubbletea and use plain text:

```go
// In ui/styles.go, add:
import "os"

// IsTTY returns true if stdout is connected to a terminal.
func IsTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
```

In `cmd/interactive.go`, wrap the bubbletea path:
```go
func runInteractive() {
    if !ui.IsTTY() {
        runInteractiveFallback() // keep old fmt-based flow as fallback
        return
    }
    // ... bubbletea flow
}
```

Keep the old `promptChoice`, `promptString`, `promptFormat` functions in a new file `cmd/interactive_fallback.go` as the non-TTY fallback path. This preserves piped input support (`echo "1" | context-gen`).

- [ ] **Step 2.6: Add terminal width adaptation to ui/styles.go**

```go
// In ui/styles.go, add:
import "golang.org/x/term"

// TermWidth returns the terminal width, defaulting to 80.
func TermWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w < 40 {
		return 80
	}
	return w
}
```

Note: `golang.org/x/term` is a lightweight dependency. Alternatively, use lipgloss's built-in width detection. If terminal width < 40, styles degrade to unbordered text by using `lipgloss.NewStyle()` (no border) instead of `BannerStyle`/`ResultsStyle`.

- [ ] **Step 3: Delete and replace interactive_test.go**

**DELETE** the entire existing `cmd/interactive_test.go`. All 11 tests (`TestPromptChoice_*`, `TestPromptString_*`, `TestPromptFormat`) call functions that have moved to `cmd/interactive_fallback.go`.

**CREATE** new `cmd/interactive_test.go` with:
- Test that `runInteractiveFallback` still works with piped input (using the old `reader` variable)
- Import the fallback functions and verify they handle valid/invalid/EOF input

The bubbletea UI models are tested in `ui/menu_test.go` (Task 8).

- [ ] **Step 4: Run all tests**

Run: `cd /home/paul/Projects/context-gen && go test ./... -v`
Expected: ALL PASS

- [ ] **Step 5: Manual test**

Run: `cd /home/paul/Projects/context-gen && go run . `
Expected: See styled banner, arrow-key menu, styled output

- [ ] **Step 6: Build and check binary**

Run: `cd /home/paul/Projects/context-gen && go build -o context-gen . && ls -lh context-gen`
Expected: Binary builds, size reasonable

- [ ] **Step 7: Commit**

```bash
git add cmd/interactive.go cmd/interactive_test.go cmd/root.go ui/
git commit -m "feat: replace plain CLI with charmbracelet TUI (menu, spinner, styled output)"
```

---

## Chunk 3: Phase 3 — License & Version

### Task 12: Add MIT License

**Files:**
- Create: `LICENSE`
- Modify: `main.go`

- [ ] **Step 1: Create LICENSE file**

Create `LICENSE` at project root with standard MIT text:

```
MIT License

Copyright (c) 2026 Paul

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

- [ ] **Step 2: Add copyright header to main.go**

Add at top of `main.go`:
```go
// Copyright (c) 2026 Paul. Licensed under the MIT License.
// See LICENSE file in the project root for full license text.
```

- [ ] **Step 3: Update version command in cmd/root.go**

Update the version case to:
```go
case "version":
    fmt.Printf("context-gen v%s\n", ui.Version)
    fmt.Println("Copyright (c) 2026 Paul")
    fmt.Println("Licensed under the MIT License")
```

- [ ] **Step 4: Update all version references to use ui.Version**

Search for any remaining `v0.1.0` strings and replace with `ui.Version` references.

- [ ] **Step 5: Verify build and test**

Run: `cd /home/paul/Projects/context-gen && go test ./... && go build ./...`
Expected: ALL PASS, BUILD SUCCESS

- [ ] **Step 6: Commit**

```bash
git add LICENSE main.go cmd/root.go
git commit -m "feat: add MIT license and bump version to v0.2.0"
```

---

### Task 13: Final verification

- [ ] **Step 1: Run full test suite**

Run: `cd /home/paul/Projects/context-gen && go test ./... -v -count=1`
Expected: ALL PASS

- [ ] **Step 2: Build final binary**

Run: `cd /home/paul/Projects/context-gen && go build -o context-gen .`
Expected: BUILD SUCCESS

- [ ] **Step 3: Test CLI commands**

```bash
./context-gen version
./context-gen help
./context-gen preview -d . -f claude
```
Expected: Styled output, correct version (v0.2.0), MIT license in version

- [ ] **Step 4: Test interactive mode**

Run: `./context-gen`
Expected: Banner box, arrow-key menu, format selector, spinner during scan, styled results

- [ ] **Step 5: Verify output line count**

Run: `./context-gen preview -d . -f claude 2>&1 | wc -l`
Expected: < 200 lines
