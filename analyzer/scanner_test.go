package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan_FileAndDirCounts(t *testing.T) {
	root := t.TempDir()

	// Create various source files.
	writeFile(t, root, "main.go", "package main")
	writeFile(t, root, "app.js", "console.log('hi')")
	writeFile(t, root, "utils.py", "def hello(): pass")

	// Create a subdirectory with files.
	mkDir(t, root, "pkg")
	writeFile(t, root, "pkg/handler.go", "package pkg")

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if result.TotalFiles != 4 {
		t.Errorf("TotalFiles = %d, want 4", result.TotalFiles)
	}
	if result.TotalDirs != 1 {
		t.Errorf("TotalDirs = %d, want 1", result.TotalDirs)
	}
}

func TestScan_ConfigDetection(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "package.json", `{"name":"test"}`)
	writeFile(t, root, "go.mod", "module test")
	writeFile(t, root, "Makefile", "all:")
	writeFile(t, root, "main.go", "package main")

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	configs := map[string]bool{}
	for _, f := range result.Files {
		if f.IsConfig {
			configs[f.Name] = true
		}
	}

	wantConfigs := []string{"package.json", "go.mod", "Makefile"}
	for _, name := range wantConfigs {
		if !configs[name] {
			t.Errorf("%s should be detected as config, but was not", name)
		}
	}

	// main.go is not a config file.
	if configs["main.go"] {
		t.Error("main.go should not be detected as config")
	}
}

func TestScan_SkipDirs(t *testing.T) {
	root := t.TempDir()

	// Create files in skip dirs.
	mkDir(t, root, "node_modules")
	writeFile(t, root, "node_modules/leftpad.js", "module.exports = 1")
	mkDir(t, root, "vendor")
	writeFile(t, root, "vendor/dep.go", "package vendor")
	mkDir(t, root, ".git")
	writeFile(t, root, ".git/config", "[core]")
	mkDir(t, root, "__pycache__")
	writeFile(t, root, "__pycache__/mod.pyc", "bytecode")

	// Create files that should be found.
	writeFile(t, root, "main.go", "package main")
	mkDir(t, root, "src")
	writeFile(t, root, "src/app.go", "package src")

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	// Only main.go and src/app.go should be present.
	if result.TotalFiles != 2 {
		t.Errorf("TotalFiles = %d, want 2 (skip dirs should be excluded)", result.TotalFiles)
	}

	// Only src should be in dirs (skip dirs should not appear).
	if result.TotalDirs != 1 {
		t.Errorf("TotalDirs = %d, want 1", result.TotalDirs)
	}

	for _, f := range result.Files {
		if f.Name == "leftpad.js" || f.Name == "dep.go" || f.Name == "config" || f.Name == "mod.pyc" {
			t.Errorf("file %s should have been skipped", f.Name)
		}
	}
}

func TestScan_HiddenFiles(t *testing.T) {
	root := t.TempDir()

	// Hidden files that are NOT config -- should be skipped.
	writeFile(t, root, ".hidden", "secret")
	writeFile(t, root, ".DS_Store", "junk")

	// Hidden config dotfiles -- should be included.
	writeFile(t, root, ".eslintrc.json", "{}")
	writeFile(t, root, ".prettierrc", "{}")

	// Regular file -- should be included.
	writeFile(t, root, "main.go", "package main")

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	names := map[string]bool{}
	for _, f := range result.Files {
		names[f.Name] = true
	}

	if names[".hidden"] {
		t.Error(".hidden should be skipped")
	}
	if names[".DS_Store"] {
		t.Error(".DS_Store should be skipped")
	}
	if !names[".eslintrc.json"] {
		t.Error(".eslintrc.json should be included (config dotfile)")
	}
	if !names[".prettierrc"] {
		t.Error(".prettierrc should be included (config dotfile)")
	}
	if !names["main.go"] {
		t.Error("main.go should be included")
	}
}

func TestScan_FileExtensions(t *testing.T) {
	root := t.TempDir()

	tests := []struct {
		filename string
		wantExt  string
	}{
		{"main.go", "go"},
		{"app.tsx", "tsx"},
		{"utils.py", "py"},
		{"Makefile", ""},
		{"config.test.ts", "ts"},
		{"styles.module.css", "css"},
	}

	for _, tt := range tests {
		writeFile(t, root, tt.filename, "content")
	}

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	extByName := map[string]string{}
	for _, f := range result.Files {
		extByName[f.Name] = f.Extension
	}

	for _, tt := range tests {
		got, ok := extByName[tt.filename]
		if !ok {
			t.Errorf("file %s not found in scan results", tt.filename)
			continue
		}
		if got != tt.wantExt {
			t.Errorf("Extension(%s) = %q, want %q", tt.filename, got, tt.wantExt)
		}
	}
}

func TestScan_EmptyDir(t *testing.T) {
	root := t.TempDir()

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if result.TotalFiles != 0 {
		t.Errorf("TotalFiles = %d, want 0", result.TotalFiles)
	}
	if result.TotalDirs != 0 {
		t.Errorf("TotalDirs = %d, want 0", result.TotalDirs)
	}
}

func TestScan_RelativePaths(t *testing.T) {
	root := t.TempDir()

	mkDir(t, root, "pkg")
	mkDir(t, root, "pkg/sub")
	writeFile(t, root, "pkg/sub/handler.go", "package sub")

	result, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	if len(result.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(result.Files))
	}

	want := filepath.Join("pkg", "sub", "handler.go")
	if result.Files[0].Path != want {
		t.Errorf("Path = %q, want %q", result.Files[0].Path, want)
	}
}

// --- helpers ---

func writeFile(t *testing.T, root, path, content string) {
	t.Helper()
	full := filepath.Join(root, path)
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile(%s): %v", path, err)
	}
}

func mkDir(t *testing.T, root, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, path), 0755); err != nil {
		t.Fatalf("mkDir(%s): %v", path, err)
	}
}
