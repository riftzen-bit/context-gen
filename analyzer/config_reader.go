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

// readEditorConfig parses a .editorconfig file and returns CodeStyle.
// Prefers values from the [*] section; falls back to the first language-specific section.
func readEditorConfig(path string) CodeStyle {
	f, err := os.Open(path)
	if err != nil {
		return CodeStyle{}
	}
	defer f.Close()

	// Parse all sections into ordered structure.
	type section struct {
		name   string
		values map[string]string
	}
	var sections []section
	var current *section

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "[") {
			sections = append(sections, section{name: line, values: make(map[string]string)})
			current = &sections[len(sections)-1]
			continue
		}

		// Lines before any section header (e.g. "root = true") are ignored.
		if current == nil {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		current.values[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	// Pick section: prefer [*], fall back to first section with values.
	var chosen map[string]string
	for _, s := range sections {
		if s.name == "[*]" {
			chosen = s.values
			break
		}
	}
	if chosen == nil {
		for _, s := range sections {
			if len(s.values) > 0 {
				chosen = s.values
				break
			}
		}
	}
	if chosen == nil {
		return CodeStyle{}
	}

	var style CodeStyle
	if v, ok := chosen["indent_style"]; ok {
		style.IndentStyle = v
	}
	if v, ok := chosen["indent_size"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			style.IndentSize = n
		}
	}
	if v, ok := chosen["max_line_length"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			style.LineLength = n
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

	// Check for ESLint config as fallback formatter for JS/TS
	for _, f := range scan.Files {
		switch f.Name {
		case ".eslintrc", ".eslintrc.js", ".eslintrc.json", ".eslintrc.yml",
			"eslint.config.js", "eslint.config.mjs", "eslint.config.ts":
			return "eslint"
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
				// Sort for deterministic output
				sort.Strings(targets)
				for _, t := range targets {
					scripts[t] = fmt.Sprintf("make %s", t)
				}
				return scripts
			}
		}
	}

	return nil
}
