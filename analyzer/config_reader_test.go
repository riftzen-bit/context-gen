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

func TestReadEditorConfig_NoGlobalSection(t *testing.T) {
	style := readEditorConfig(filepath.Join(testdataDir(), "editorconfig_no_global"))

	if style.IndentStyle != "space" {
		t.Errorf("IndentStyle = %q, want %q", style.IndentStyle, "space")
	}
	if style.IndentSize != 2 {
		t.Errorf("IndentSize = %d, want %d", style.IndentSize, 2)
	}
	if style.LineLength != 100 {
		t.Errorf("LineLength = %d, want %d", style.LineLength, 100)
	}
}

func TestReadEditorConfig_OnlyLangSection(t *testing.T) {
	style := readEditorConfig(filepath.Join(testdataDir(), "editorconfig_only_lang"))

	if style.IndentStyle != "space" {
		t.Errorf("IndentStyle = %q, want %q", style.IndentStyle, "space")
	}
	if style.IndentSize != 4 {
		t.Errorf("IndentSize = %d, want %d", style.IndentSize, 4)
	}
	if style.LineLength != 120 {
		t.Errorf("LineLength = %d, want %d", style.LineLength, 120)
	}
}

func TestReadEditorConfig_GlobalSectionTakesPriority(t *testing.T) {
	// editorconfig_sample has [*] with tab/4/120 — verify [*] values are used
	style := readEditorConfig(filepath.Join(testdataDir(), "editorconfig_sample"))

	if style.IndentStyle != "tab" {
		t.Errorf("IndentStyle = %q, want %q (from [*] section)", style.IndentStyle, "tab")
	}
	if style.IndentSize != 4 {
		t.Errorf("IndentSize = %d, want %d (from [*] section)", style.IndentSize, 4)
	}
	if style.LineLength != 120 {
		t.Errorf("LineLength = %d, want %d (from [*] section)", style.LineLength, 120)
	}
}
