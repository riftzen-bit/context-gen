package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectLanguages(t *testing.T) {
	tests := []struct {
		name      string
		files     []FileInfo
		wantLangs map[string]bool
		wantFirst string // primary language (highest count)
	}{
		{
			name: "Go only",
			files: []FileInfo{
				{Name: "main.go", Extension: "go"},
				{Name: "handler.go", Extension: "go"},
				{Name: "utils.go", Extension: "go"},
			},
			wantLangs: map[string]bool{"Go": true},
			wantFirst: "Go",
		},
		{
			name: "mixed JS and Go",
			files: []FileInfo{
				{Name: "main.go", Extension: "go"},
				{Name: "app.js", Extension: "js"},
				{Name: "utils.js", Extension: "js"},
				{Name: "index.js", Extension: "js"},
			},
			wantLangs: map[string]bool{"Go": true, "JavaScript": true},
			wantFirst: "JavaScript",
		},
		{
			name: "config files excluded",
			files: []FileInfo{
				{Name: "main.go", Extension: "go"},
				{Name: "package.json", Extension: "json", IsConfig: true},
				{Name: "go.mod", Extension: "mod", IsConfig: true},
			},
			wantLangs: map[string]bool{"Go": true},
		},
		{
			name: "files with no extension excluded",
			files: []FileInfo{
				{Name: "main.py", Extension: "py"},
				{Name: "Makefile", Extension: ""},
			},
			wantLangs: map[string]bool{"Python": true},
		},
		{
			name: "TypeScript React",
			files: []FileInfo{
				{Name: "App.tsx", Extension: "tsx"},
				{Name: "index.ts", Extension: "ts"},
			},
			wantLangs: map[string]bool{"TypeScript (React)": true, "TypeScript": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scan := &ScanResult{Files: tt.files}
			langs := detectLanguages(scan)

			detected := map[string]bool{}
			for _, l := range langs {
				detected[l.Name] = true
			}

			for want := range tt.wantLangs {
				if !detected[want] {
					t.Errorf("expected language %q not detected", want)
				}
			}

			for name := range detected {
				if !tt.wantLangs[name] {
					t.Errorf("unexpected language %q detected", name)
				}
			}

			if tt.wantFirst != "" && len(langs) > 0 {
				if langs[0].Name != tt.wantFirst {
					t.Errorf("primary language = %q, want %q", langs[0].Name, tt.wantFirst)
				}
			}
		})
	}
}

func TestDetectLanguages_Percentages(t *testing.T) {
	scan := &ScanResult{
		Files: []FileInfo{
			{Name: "a.go", Extension: "go"},
			{Name: "b.go", Extension: "go"},
			{Name: "c.go", Extension: "go"},
			{Name: "d.py", Extension: "py"},
		},
	}

	langs := detectLanguages(scan)

	langMap := map[string]Language{}
	for _, l := range langs {
		langMap[l.Name] = l
	}

	goLang, ok := langMap["Go"]
	if !ok {
		t.Fatal("Go language not detected")
	}
	if goLang.FileCount != 3 {
		t.Errorf("Go FileCount = %d, want 3", goLang.FileCount)
	}
	if goLang.Percentage != 75.0 {
		t.Errorf("Go Percentage = %f, want 75.0", goLang.Percentage)
	}

	pyLang, ok := langMap["Python"]
	if !ok {
		t.Fatal("Python language not detected")
	}
	if pyLang.Percentage != 25.0 {
		t.Errorf("Python Percentage = %f, want 25.0", pyLang.Percentage)
	}
}

func TestDetectFromConfigs_GoProject(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "go.mod", "module example.com/test\ngo 1.21")

	scan := &ScanResult{
		Files: []FileInfo{
			{Name: "go.mod", Path: "go.mod", Extension: "mod", IsConfig: true},
			{Name: "main.go", Path: "main.go", Extension: "go"},
		},
	}

	info := &ProjectInfo{PackageManagers: make(map[string]string)}
	detectFromConfigs(root, scan, info)

	if !contains(info.BuildTools, "Go Modules") {
		t.Error("expected Go Modules in BuildTools")
	}
}

func TestDetectFromConfigs_NodeProject(t *testing.T) {
	root := t.TempDir()

	pkgJSON := `{
		"dependencies": {
			"react": "^18.0.0",
			"next": "^14.0.0"
		},
		"devDependencies": {
			"eslint": "^8.0.0",
			"vitest": "^1.0.0",
			"tailwindcss": "^3.0.0"
		}
	}`
	writeFile(t, root, "package.json", pkgJSON)
	writeFile(t, root, "pnpm-lock.yaml", "lockfileVersion: 9")

	scan := &ScanResult{
		Files: []FileInfo{
			{Name: "package.json", Path: "package.json", Extension: "json", IsConfig: true},
		},
	}

	info := &ProjectInfo{PackageManagers: make(map[string]string)}
	detectFromConfigs(root, scan, info)

	if info.PackageManagers["js"] != "pnpm" {
		t.Errorf("PackageManagers[js] = %q, want pnpm", info.PackageManagers["js"])
	}
	if !contains(info.Frameworks, "React") {
		t.Error("expected React in Frameworks")
	}
	if !contains(info.Frameworks, "Next.js") {
		t.Error("expected Next.js in Frameworks")
	}
	if !contains(info.Frameworks, "Tailwind CSS") {
		t.Error("expected Tailwind CSS in Frameworks")
	}
	if !contains(info.Linters, "ESLint") {
		t.Error("expected ESLint in Linters")
	}
	if !contains(info.TestTools, "Vitest") {
		t.Error("expected Vitest in TestTools")
	}
}

func TestDetectFromConfigs_PythonPoetry(t *testing.T) {
	root := t.TempDir()

	pyproject := `[tool.poetry]
name = "myproject"
version = "0.1.0"

[tool.poetry.dependencies]
python = "^3.11"

[tool.pytest.ini_options]
testpaths = ["tests"]

[tool.ruff]
line-length = 88
`
	writeFile(t, root, "pyproject.toml", pyproject)

	scan := &ScanResult{
		Files: []FileInfo{
			{Name: "pyproject.toml", Path: "pyproject.toml", Extension: "toml", IsConfig: true},
		},
	}

	info := &ProjectInfo{PackageManagers: make(map[string]string)}
	detectFromConfigs(root, scan, info)

	if info.PackageManagers["python"] != "poetry" {
		t.Errorf("PackageManagers[python] = %q, want poetry", info.PackageManagers["python"])
	}
	if !contains(info.BuildTools, "pyproject") {
		t.Error("expected pyproject in BuildTools")
	}
	if !contains(info.TestTools, "pytest") {
		t.Error("expected pytest in TestTools")
	}
	if !contains(info.Linters, "ruff") {
		t.Error("expected ruff in Linters")
	}
}

func TestDetectFromConfigs_MultiLanguage(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "go.mod", "module test\ngo 1.21")
	writeFile(t, root, "package.json", `{"dependencies":{"express":"^4.0.0"}}`)
	writeFile(t, root, "Makefile", "all:\n\tgo build ./...")

	scan := &ScanResult{
		Files: []FileInfo{
			{Name: "go.mod", Path: "go.mod", Extension: "mod", IsConfig: true},
			{Name: "package.json", Path: "package.json", Extension: "json", IsConfig: true},
			{Name: "Makefile", Path: "Makefile", Extension: "", IsConfig: true},
		},
	}

	info := &ProjectInfo{PackageManagers: make(map[string]string)}
	detectFromConfigs(root, scan, info)

	if !contains(info.BuildTools, "Go Modules") {
		t.Error("expected Go Modules in BuildTools")
	}
	if !contains(info.BuildTools, "Make") {
		t.Error("expected Make in BuildTools")
	}
	if !contains(info.Frameworks, "Express") {
		t.Error("expected Express in Frameworks")
	}
	if info.PackageManagers["js"] == "" {
		t.Error("expected js package manager to be set")
	}
}

func TestDetectCI(t *testing.T) {
	tests := []struct {
		name       string
		dirs       []string
		files      []FileInfo
		wantCI     bool
		wantProvider string
	}{
		{
			name:       "GitHub Actions",
			dirs:       []string{".github", ".github/workflows"},
			wantCI:     true,
			wantProvider: "GitHub Actions",
		},
		{
			name:       "Travis CI",
			files:      []FileInfo{{Name: ".travis.yml", Path: ".travis.yml"}},
			wantCI:     true,
			wantProvider: "Travis CI",
		},
		{
			name:       "GitLab CI from file",
			files:      []FileInfo{{Name: ".gitlab-ci.yml", Path: ".gitlab-ci.yml"}},
			wantCI:     true,
			wantProvider: "GitLab CI",
		},
		{
			name:       "No CI",
			dirs:       []string{"src"},
			wantCI:     false,
			wantProvider: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scan := &ScanResult{Dirs: tt.dirs, Files: tt.files}
			info := &ProjectInfo{PackageManagers: make(map[string]string)}
			detectCI(scan, info)

			if info.HasCI != tt.wantCI {
				t.Errorf("HasCI = %v, want %v", info.HasCI, tt.wantCI)
			}
			if info.CIProvider != tt.wantProvider {
				t.Errorf("CIProvider = %q, want %q", info.CIProvider, tt.wantProvider)
			}
		})
	}
}

func TestDetectDocker(t *testing.T) {
	tests := []struct {
		name       string
		files      []FileInfo
		wantDocker bool
	}{
		{
			name:       "Dockerfile present",
			files:      []FileInfo{{Name: "Dockerfile"}},
			wantDocker: true,
		},
		{
			name:       "docker-compose.yml",
			files:      []FileInfo{{Name: "docker-compose.yml"}},
			wantDocker: true,
		},
		{
			name:       "no docker",
			files:      []FileInfo{{Name: "main.go"}},
			wantDocker: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scan := &ScanResult{Files: tt.files}
			info := &ProjectInfo{PackageManagers: make(map[string]string)}
			detectDocker(scan, info)

			if info.HasDocker != tt.wantDocker {
				t.Errorf("HasDocker = %v, want %v", info.HasDocker, tt.wantDocker)
			}
		})
	}
}

func TestDetect_Integration(t *testing.T) {
	root := t.TempDir()

	// Build a realistic Go project structure.
	writeFile(t, root, "go.mod", "module example.com/myapp\ngo 1.21")
	writeFile(t, root, "main.go", `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`)
	mkDir(t, root, "pkg")
	writeFile(t, root, "pkg/handler.go", `package pkg

import "fmt"

func HandleRequest() error {
	if err := doWork(); err != nil {
		return fmt.Errorf("handle: %w", err)
	}
	return nil
}

func doWork() error {
	return nil
}
`)
	writeFile(t, root, "pkg/handler_test.go", `package pkg

import "testing"

func TestHandleRequest(t *testing.T) {
}
`)
	mkDir(t, root, ".github")
	mkDir(t, root, ".github/workflows")
	writeFile(t, root, "Makefile", "all:\n\tgo build ./...")
	writeFile(t, root, "Dockerfile", "FROM golang:1.21")

	scan, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	info, err := Detect(root, scan)
	if err != nil {
		t.Fatalf("Detect() error: %v", err)
	}

	// Language detection.
	if len(info.Languages) == 0 {
		t.Fatal("expected at least one language")
	}
	if info.Languages[0].Name != "Go" {
		t.Errorf("primary language = %q, want Go", info.Languages[0].Name)
	}

	// Build tools.
	if !contains(info.BuildTools, "Go Modules") {
		t.Error("expected Go Modules in BuildTools")
	}
	if !contains(info.BuildTools, "Make") {
		t.Error("expected Make in BuildTools")
	}

	// CI.
	if !info.HasCI {
		t.Error("expected HasCI = true")
	}
	if info.CIProvider != "GitHub Actions" {
		t.Errorf("CIProvider = %q, want GitHub Actions", info.CIProvider)
	}

	// Docker.
	if !info.HasDocker {
		t.Error("expected HasDocker = true")
	}

	// Structure.
	if info.Structure.TotalFiles == 0 {
		t.Error("expected TotalFiles > 0")
	}
	foundPkgDir := false
	for _, d := range info.Structure.TopLevelDirs {
		if d == "pkg" {
			foundPkgDir = true
		}
	}
	if !foundPkgDir {
		t.Error("expected 'pkg' in TopLevelDirs")
	}

	// Entry point.
	foundMain := false
	for _, ep := range info.Structure.EntryPoints {
		if ep == "main.go" {
			foundMain = true
		}
	}
	if foundMain == false {
		t.Error("expected main.go as entry point")
	}
}

func TestBuildStructure(t *testing.T) {
	scan := &ScanResult{
		TotalFiles: 5,
		TotalDirs:  3,
		Files: []FileInfo{
			{Name: "main.go", Path: "main.go"},
			{Name: "index.ts", Path: "src/index.ts"},
			{Name: "handler.go", Path: "pkg/handler.go"},
			{Name: "package.json", Path: "package.json", IsConfig: true},
			{Name: "go.mod", Path: "go.mod", IsConfig: true},
		},
		Dirs: []string{
			"src",
			"pkg",
			filepath.Join("pkg", "sub"),
		},
	}

	ds := buildStructure(scan)

	if ds.TotalFiles != 5 {
		t.Errorf("TotalFiles = %d, want 5", ds.TotalFiles)
	}
	if ds.TotalDirs != 3 {
		t.Errorf("TotalDirs = %d, want 3", ds.TotalDirs)
	}

	// Top level dirs should be src and pkg (not pkg/sub).
	topLevelSet := map[string]bool{}
	for _, d := range ds.TopLevelDirs {
		topLevelSet[d] = true
	}
	if !topLevelSet["src"] {
		t.Error("expected 'src' in TopLevelDirs")
	}
	if !topLevelSet["pkg"] {
		t.Error("expected 'pkg' in TopLevelDirs")
	}

	// Entry points.
	epSet := map[string]bool{}
	for _, ep := range ds.EntryPoints {
		epSet[ep] = true
	}
	if !epSet["main.go"] {
		t.Error("expected main.go as entry point")
	}
	if !epSet["src/index.ts"] {
		t.Error("expected src/index.ts as entry point")
	}

	// Config files.
	cfgSet := map[string]bool{}
	for _, cf := range ds.ConfigFiles {
		cfgSet[cf] = true
	}
	if !cfgSet["package.json"] {
		t.Error("expected package.json in ConfigFiles")
	}
	if !cfgSet["go.mod"] {
		t.Error("expected go.mod in ConfigFiles")
	}
}

func TestAppendUnique(t *testing.T) {
	slice := []string{"a", "b"}
	slice = appendUnique(slice, "c")
	if len(slice) != 3 {
		t.Errorf("expected 3 items, got %d", len(slice))
	}

	// Adding duplicate should not increase length.
	slice = appendUnique(slice, "b")
	if len(slice) != 3 {
		t.Errorf("expected 3 items after adding duplicate, got %d", len(slice))
	}
}

func TestMergeMaps(t *testing.T) {
	a := map[string]string{"x": "1", "y": "2"}
	b := map[string]string{"y": "3", "z": "4"}

	result := mergeMaps(a, b)

	if len(result) != 3 {
		t.Errorf("expected 3 entries, got %d", len(result))
	}
	// b should override a for key "y".
	if result["y"] != "3" {
		t.Errorf("y = %q, want 3", result["y"])
	}
	if result["z"] != "4" {
		t.Errorf("z = %q, want 4", result["z"])
	}
}

func TestDetectFromPackageJSON_PackageManagerDetection(t *testing.T) {
	tests := []struct {
		name     string
		lockFile string
		wantPM   string
	}{
		{"bun", "bun.lockb", "bun"},
		{"pnpm", "pnpm-lock.yaml", "pnpm"},
		{"yarn", "yarn.lock", "yarn"},
		{"npm default", "", "npm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			writeFile(t, root, "package.json", `{"dependencies":{}}`)
			if tt.lockFile != "" {
				writeFile(t, root, tt.lockFile, "lock content")
			}

			info := &ProjectInfo{PackageManagers: make(map[string]string)}
			detectFromPackageJSON(root, "package.json", info)

			if info.PackageManagers["js"] != tt.wantPM {
				t.Errorf("PackageManagers[js] = %q, want %q", info.PackageManagers["js"], tt.wantPM)
			}
		})
	}
}

func TestDetect_PopulatesName(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module github.com/test/myproject\n\ngo 1.21\n")
	writeFile(t, dir, "main.go", "package main\n")

	scan, _ := Scan(dir)
	info, err := Detect(dir, scan)
	if err != nil {
		t.Fatal(err)
	}

	if info.Name != "myproject" {
		t.Errorf("Name = %q, want %q", info.Name, "myproject")
	}
}

func TestDetect_PopulatesCodeStyle(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module github.com/test/myproject\n\ngo 1.21\n")
	writeFile(t, dir, "main.go", "package main\n")

	scan, _ := Scan(dir)
	info, err := Detect(dir, scan)
	if err != nil {
		t.Fatal(err)
	}

	// Go project should get gofmt as default formatter
	if info.CodeStyle.Formatter != "gofmt" {
		t.Errorf("CodeStyle.Formatter = %q, want %q", info.CodeStyle.Formatter, "gofmt")
	}
}

// --- helpers ---

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// writeFile and mkDir are defined in scanner_test.go (same package).
// These helper stubs are here only for reference if tests are run in isolation.
// Since Go test builds all _test.go files in the same package together,
// the helpers from scanner_test.go are available here.
func init() {
	// Ensure test temp dirs get cleaned up via t.TempDir().
	_ = os.TempDir()
}
