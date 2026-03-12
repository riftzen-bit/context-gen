package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseGitIgnore_BasicPatterns(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".gitignore", `# Build output
dist/
build/
*.log

# Dependencies
node_modules/

# IDE
.idea/
*.swp

# Environment
.env
.env.local
`)

	gi := ParseGitIgnore(root)

	if !gi.HasPatterns() {
		t.Fatal("expected patterns to be loaded")
	}

	tests := []struct {
		path    string
		isDir   bool
		ignored bool
	}{
		{"dist", true, true},
		{"build", true, true},
		{"node_modules", true, true},
		{".idea", true, true},
		{"debug.log", false, true},
		{"logs/error.log", false, true},
		{"main.go", false, false},
		{"src/app.ts", false, false},
		{".env", false, true},
		{".env.local", false, true},
		{"config.env", false, false}, // not .env
		{"app.swp", false, true},
		// dir-only patterns should not match files
		{"dist", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := gi.IsIgnored(tt.path, tt.isDir)
			if got != tt.ignored {
				t.Errorf("IsIgnored(%q, isDir=%v) = %v, want %v", tt.path, tt.isDir, got, tt.ignored)
			}
		})
	}
}

func TestParseGitIgnore_Negation(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".gitignore", `*.log
!important.log
`)

	gi := ParseGitIgnore(root)

	if gi.IsIgnored("important.log", false) {
		t.Error("important.log should NOT be ignored (negated)")
	}
	if !gi.IsIgnored("debug.log", false) {
		t.Error("debug.log should be ignored")
	}
	if !gi.IsIgnored("error.log", false) {
		t.Error("error.log should be ignored")
	}
}

func TestParseGitIgnore_AnchoredPatterns(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".gitignore", `# Only root-level tmp, not nested
/tmp
# Nested pattern
src/generated/
`)

	gi := ParseGitIgnore(root)

	tests := []struct {
		path    string
		isDir   bool
		ignored bool
	}{
		{"tmp", true, true},
		{"tmp", false, true},
		{"sub/tmp", true, false},  // /tmp only matches at root
		{"sub/tmp", false, false}, // /tmp only matches at root
		{"src/generated", true, true},
		{"other/generated", true, false}, // anchored to src/
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := gi.IsIgnored(tt.path, tt.isDir)
			if got != tt.ignored {
				t.Errorf("IsIgnored(%q, isDir=%v) = %v, want %v", tt.path, tt.isDir, got, tt.ignored)
			}
		})
	}
}

func TestParseGitIgnore_DoubleStarPatterns(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".gitignore", `**/logs
docs/**/*.temp
`)

	gi := ParseGitIgnore(root)

	tests := []struct {
		path    string
		isDir   bool
		ignored bool
	}{
		{"logs", true, true},
		{"sub/logs", true, true},
		{"deep/sub/logs", true, true},
		{"docs/draft.temp", false, true},
		{"docs/sub/draft.temp", false, true},
		{"other/draft.temp", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := gi.IsIgnored(tt.path, tt.isDir)
			if got != tt.ignored {
				t.Errorf("IsIgnored(%q, isDir=%v) = %v, want %v", tt.path, tt.isDir, got, tt.ignored)
			}
		})
	}
}

func TestParseGitIgnore_CommentsAndEmptyLines(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".gitignore", `
# This is a comment

*.log

# Another comment
`)

	gi := ParseGitIgnore(root)

	if !gi.IsIgnored("test.log", false) {
		t.Error("*.log should match")
	}
}

func TestParseGitIgnore_MissingFile(t *testing.T) {
	root := t.TempDir()

	gi := ParseGitIgnore(root)

	if gi.HasPatterns() {
		t.Error("no patterns should be loaded for missing .gitignore")
	}
	if gi.IsIgnored("anything", false) {
		t.Error("nothing should be ignored without .gitignore")
	}
}

func TestParseGitIgnore_EmptyFile(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, ".gitignore", "")

	gi := ParseGitIgnore(root)

	if gi.HasPatterns() {
		t.Error("no patterns should be loaded from empty .gitignore")
	}
}

func TestParseGitIgnore_Integration_WithScan(t *testing.T) {
	root := t.TempDir()

	// Create project structure
	writeFile(t, root, ".gitignore", "tmp/\n*.log\n")
	writeFile(t, root, "main.go", "package main\n")
	mkDir(t, root, "src")
	writeFile(t, root, "src/app.go", "package src\n")
	mkDir(t, root, "tmp")
	writeFile(t, root, "tmp/cache.txt", "cached")
	writeFile(t, root, "debug.log", "log output")

	// Scan should respect .gitignore
	gi := ParseGitIgnore(root)

	// Verify patterns work
	if !gi.IsIgnored("tmp", true) {
		t.Error("tmp/ should be ignored")
	}
	if !gi.IsIgnored("debug.log", false) {
		t.Error("debug.log should be ignored")
	}
	if gi.IsIgnored("main.go", false) {
		t.Error("main.go should NOT be ignored")
	}
	if gi.IsIgnored("src/app.go", false) {
		t.Error("src/app.go should NOT be ignored")
	}
}

func TestParseLine_EdgeCases(t *testing.T) {
	tests := []struct {
		line string
		want *ignorePattern
	}{
		{"", nil},
		{"# comment", nil},
		{"  # comment with leading space", nil},
		{"*.log", &ignorePattern{pattern: "*.log"}},
		{"!keep.log", &ignorePattern{pattern: "keep.log", negate: true}},
		{"build/", &ignorePattern{pattern: "build", dirOnly: true}},
		{"/root-only", &ignorePattern{pattern: "root-only", anchored: true}},
		{"src/dist/", &ignorePattern{pattern: "src/dist", dirOnly: true, anchored: true}},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := parseLine(tt.line)

			if tt.want == nil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected non-nil pattern")
			}
			if got.pattern != tt.want.pattern {
				t.Errorf("pattern = %q, want %q", got.pattern, tt.want.pattern)
			}
			if got.negate != tt.want.negate {
				t.Errorf("negate = %v, want %v", got.negate, tt.want.negate)
			}
			if got.dirOnly != tt.want.dirOnly {
				t.Errorf("dirOnly = %v, want %v", got.dirOnly, tt.want.dirOnly)
			}
			if got.anchored != tt.want.anchored {
				t.Errorf("anchored = %v, want %v", got.anchored, tt.want.anchored)
			}
		})
	}
}

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		pattern string
		str     string
		match   bool
	}{
		{"*.log", "debug.log", true},
		{"*.log", "error.log", true},
		{"*.log", "app.txt", false},
		{"*.log", "dir/debug.log", false}, // * doesn't match /
		{"test", "test", true},
		{"test", "testing", false},
		{"test?", "test1", true},
		{"test?", "testAB", false},
		{"*.ts", "app.ts", true},
		{"*.ts", "app.tsx", false},
	}

	for _, tt := range tests {
		name := tt.pattern + "_vs_" + tt.str
		t.Run(name, func(t *testing.T) {
			got := matchGlob(tt.pattern, tt.str)
			if got != tt.match {
				t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.str, got, tt.match)
			}
		})
	}
}

// writeFile and mkDir helpers are defined in scanner_test.go (same package).
// They are available here because Go test builds all _test.go files together.
func init() {
	_ = os.TempDir()
	_ = filepath.Join
}
