package analyzer

import (
	"path/filepath"
	"testing"
)

func TestClassifyName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"my-component", "kebab-case"},
		{"my-long-name", "kebab-case"},
		{"my_module", "snake_case"},
		{"my_long_name", "snake_case"},
		{"MyComponent", "PascalCase"},
		{"MyLongName", "PascalCase"},
		{"myFunction", "camelCase"},
		{"myLongName", "camelCase"},
		{"main", ""},          // single lowercase word
		{"", ""},              // empty string
		{"ALLCAPS", ""},       // all caps, no lowercase after first char... actually has uppercase so check
		{"alllower", ""},      // all lowercase, no separators, no uppercase
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyName(tt.name)
			if got != tt.want {
				t.Errorf("classifyName(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestClassifyName_MixedSeparators(t *testing.T) {
	// Names with both - and _ return empty (no clear style).
	got := classifyName("my-mixed_name")
	if got != "" {
		t.Errorf("classifyName(%q) = %q, want empty", "my-mixed_name", got)
	}
}

func TestDetectFileNaming(t *testing.T) {
	tests := []struct {
		name      string
		files     []FileInfo
		wantStyle string
		wantNil   bool
	}{
		{
			name: "kebab-case files",
			files: []FileInfo{
				{Name: "my-component.ts", Extension: "ts"},
				{Name: "auth-service.ts", Extension: "ts"},
				{Name: "user-profile.ts", Extension: "ts"},
			},
			wantStyle: "kebab-case",
		},
		{
			name: "snake_case files",
			files: []FileInfo{
				{Name: "my_module.py", Extension: "py"},
				{Name: "auth_service.py", Extension: "py"},
				{Name: "user_profile.py", Extension: "py"},
			},
			wantStyle: "snake_case",
		},
		{
			name: "PascalCase files",
			files: []FileInfo{
				{Name: "MyComponent.tsx", Extension: "tsx"},
				{Name: "AuthService.ts", Extension: "ts"},
				{Name: "UserProfile.tsx", Extension: "tsx"},
			},
			wantStyle: "PascalCase",
		},
		{
			name: "camelCase files",
			files: []FileInfo{
				{Name: "myComponent.ts", Extension: "ts"},
				{Name: "authService.ts", Extension: "ts"},
				{Name: "userProfile.ts", Extension: "ts"},
			},
			wantStyle: "camelCase",
		},
		{
			name: "single-word names only",
			files: []FileInfo{
				{Name: "main.go", Extension: "go"},
				{Name: "utils.go", Extension: "go"},
				{Name: "handler.go", Extension: "go"},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conventions := detectFileNaming(tt.files)

			if tt.wantNil {
				if conventions != nil {
					t.Errorf("expected nil conventions, got %v", conventions)
				}
				return
			}

			if len(conventions) == 0 {
				t.Fatal("expected at least one convention")
			}

			conv := conventions[0]
			if conv.Category != "naming" {
				t.Errorf("Category = %q, want naming", conv.Category)
			}
			if conv.Confidence < 0.5 {
				t.Errorf("Confidence = %f, want >= 0.5", conv.Confidence)
			}

			wantDesc := "Files use " + tt.wantStyle + " naming"
			if conv.Description != wantDesc {
				t.Errorf("Description = %q, want %q", conv.Description, wantDesc)
			}
		})
	}
}

func TestDetectTestStructure_CoLocated(t *testing.T) {
	scan := &ScanResult{
		Files: []FileInfo{
			{Name: "handler.go", Path: "pkg/handler.go", Extension: "go"},
			{Name: "handler_test.go", Path: "pkg/handler_test.go", Extension: "go"},
			{Name: "service.go", Path: "pkg/service.go", Extension: "go"},
			{Name: "service_test.go", Path: "pkg/service_test.go", Extension: "go"},
		},
	}

	conventions := detectTestStructure(scan)

	if len(conventions) == 0 {
		t.Fatal("expected test structure convention")
	}

	conv := conventions[0]
	if conv.Description != "Tests are co-located alongside source files" {
		t.Errorf("Description = %q, want co-located", conv.Description)
	}
	if conv.Confidence != 1.0 {
		t.Errorf("Confidence = %f, want 1.0", conv.Confidence)
	}
}

func TestDetectTestStructure_SeparateDir(t *testing.T) {
	scan := &ScanResult{
		Files: []FileInfo{
			{Name: "handler.go", Path: "pkg/handler.go", Extension: "go"},
			{Name: "handler_test.go", Path: filepath.Join("tests", "handler_test.go"), Extension: "go"},
			{Name: "service_test.go", Path: filepath.Join("tests", "service_test.go"), Extension: "go"},
			{Name: "app.test.ts", Path: filepath.Join("__tests__", "app.test.ts"), Extension: "ts"},
		},
	}

	conventions := detectTestStructure(scan)

	if len(conventions) == 0 {
		t.Fatal("expected test structure convention")
	}

	conv := conventions[0]
	if conv.Description != "Tests are in separate test directories" {
		t.Errorf("Description = %q, want separate directories", conv.Description)
	}
	if conv.Confidence != 1.0 {
		t.Errorf("Confidence = %f, want 1.0", conv.Confidence)
	}
}

func TestDetectTestStructure_JSPatterns(t *testing.T) {
	scan := &ScanResult{
		Files: []FileInfo{
			{Name: "app.test.js", Path: "src/app.test.js", Extension: "js"},
			{Name: "utils.spec.ts", Path: "src/utils.spec.ts", Extension: "ts"},
		},
	}

	conventions := detectTestStructure(scan)

	if len(conventions) == 0 {
		t.Fatal("expected test structure convention")
	}

	conv := conventions[0]
	if conv.Description != "Tests are co-located alongside source files" {
		t.Errorf("Description = %q, want co-located", conv.Description)
	}
}

func TestDetectTestStructure_PythonPattern(t *testing.T) {
	scan := &ScanResult{
		Files: []FileInfo{
			{Name: "test_handler.py", Path: filepath.Join("tests", "test_handler.py"), Extension: "py"},
			{Name: "test_service.py", Path: filepath.Join("tests", "test_service.py"), Extension: "py"},
		},
	}

	conventions := detectTestStructure(scan)

	if len(conventions) == 0 {
		t.Fatal("expected test structure convention")
	}

	conv := conventions[0]
	if conv.Description != "Tests are in separate test directories" {
		t.Errorf("Description = %q, want separate directories", conv.Description)
	}
}

func TestDetectTestStructure_NoTests(t *testing.T) {
	scan := &ScanResult{
		Files: []FileInfo{
			{Name: "handler.go", Path: "pkg/handler.go", Extension: "go"},
			{Name: "service.go", Path: "pkg/service.go", Extension: "go"},
		},
	}

	conventions := detectTestStructure(scan)

	if conventions != nil {
		t.Errorf("expected nil conventions for no test files, got %v", conventions)
	}
}

func TestDetectFunctionNaming_Go(t *testing.T) {
	contents := []fileContent{
		{
			info: FileInfo{Extension: "go"},
			content: `package pkg

func HandleRequest() error {
	return nil
}

func doWork() error {
	return nil
}

func (s *Server) Start() error {
	return nil
}

func (s *Server) listen() {}
`,
		},
	}

	conventions := detectFunctionNaming(contents)

	if len(conventions) == 0 {
		t.Fatal("expected Go naming convention")
	}

	found := false
	for _, c := range conventions {
		if c.Description == "Go functions follow standard naming: PascalCase for exported, camelCase for unexported" {
			found = true
			if c.Confidence != 1.0 {
				t.Errorf("Confidence = %f, want 1.0", c.Confidence)
			}
		}
	}
	if !found {
		t.Error("expected Go standard naming convention")
	}
}

func TestDetectFunctionNaming_JS(t *testing.T) {
	contents := []fileContent{
		{
			info: FileInfo{Extension: "ts"},
			content: `function handleRequest() {
	return null;
}

const processData = (data) => {
	return data;
}

let doSomething = () => {};
`,
		},
	}

	conventions := detectFunctionNaming(contents)

	found := false
	for _, c := range conventions {
		if c.Description == "Functions use camelCase naming" {
			found = true
		}
	}
	if !found {
		t.Error("expected camelCase naming convention for JS/TS")
	}
}

func TestDetectFunctionNaming_Python(t *testing.T) {
	contents := []fileContent{
		{
			info: FileInfo{Extension: "py"},
			content: `def handle_request():
    pass

def process_data():
    pass

def do_something():
    pass
`,
		},
	}

	conventions := detectFunctionNaming(contents)

	found := false
	for _, c := range conventions {
		if c.Description == "Python functions use snake_case naming" {
			found = true
			if c.Confidence < 0.5 {
				t.Errorf("Confidence = %f, want >= 0.5", c.Confidence)
			}
		}
	}
	if !found {
		t.Error("expected Python snake_case naming convention")
	}
}

func TestDetectErrorHandling_Go(t *testing.T) {
	contents := []fileContent{
		{
			info: FileInfo{Extension: "go"},
			content: `package pkg

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("not found")

func doWork() error {
	if err != nil {
		return fmt.Errorf("doWork: %w", err)
	}
	if err != nil {
		return err
	}
	return nil
}
`,
		},
	}

	conventions := detectErrorHandling(contents)

	descs := map[string]bool{}
	for _, c := range conventions {
		descs[c.Description] = true
	}

	if !descs["Go uses standard if err != nil error checking"] {
		t.Error("expected err != nil convention")
	}
	if !descs["Errors are wrapped with fmt.Errorf and %w for context"] {
		t.Error("expected error wrapping convention")
	}
	if !descs["Sentinel errors are defined with errors.New or fmt.Errorf"] {
		t.Error("expected sentinel error convention")
	}
}

func TestDetectErrorHandling_JS_AsyncAwait(t *testing.T) {
	contents := []fileContent{
		{
			info: FileInfo{Extension: "ts"},
			content: `async function fetchData() {
	try {
		const data = await fetch(url);
	} catch (err) {
		console.error(err);
	}
}

async function processData() {
	try {
		await process();
	} catch (e) {
		throw e;
	}
}
`,
		},
	}

	conventions := detectErrorHandling(contents)

	found := false
	for _, c := range conventions {
		if c.Description == "async/await with try/catch for error handling" {
			found = true
		}
	}
	if !found {
		t.Error("expected async/await try/catch convention")
	}
}

func TestDetectImportStyle_GoGrouped(t *testing.T) {
	contents := []fileContent{
		{
			info: FileInfo{Extension: "go"},
			content: `package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)
`,
		},
	}

	conventions := detectImportStyle(contents)

	found := false
	for _, c := range conventions {
		if c.Category == "imports" && c.Confidence > 0.5 {
			found = true
		}
	}
	if !found {
		t.Error("expected Go import style convention")
	}
}

func TestFilterByExt(t *testing.T) {
	contents := []fileContent{
		{info: FileInfo{Extension: "go"}},
		{info: FileInfo{Extension: "ts"}},
		{info: FileInfo{Extension: "tsx"}},
		{info: FileInfo{Extension: "py"}},
		{info: FileInfo{Extension: "js"}},
	}

	goFiles := filterByExt(contents, "go")
	if len(goFiles) != 1 {
		t.Errorf("filterByExt(go) returned %d files, want 1", len(goFiles))
	}

	jsFiles := filterByExt(contents, "js", "ts", "tsx")
	if len(jsFiles) != 3 {
		t.Errorf("filterByExt(js,ts,tsx) returned %d files, want 3", len(jsFiles))
	}

	empty := filterByExt(contents, "rs")
	if len(empty) != 0 {
		t.Errorf("filterByExt(rs) returned %d files, want 0", len(empty))
	}
}

func TestDedup(t *testing.T) {
	input := []string{"a", "b", "a", "c", "b"}
	result := dedup(input)

	if len(result) != 3 {
		t.Errorf("dedup returned %d items, want 3", len(result))
	}

	// Verify order is preserved.
	want := []string{"a", "b", "c"}
	for i, v := range result {
		if v != want[i] {
			t.Errorf("dedup[%d] = %q, want %q", i, v, want[i])
		}
	}
}

func TestSelectSourceFiles(t *testing.T) {
	// Should exclude config files, non-source extensions, empty files, and oversized files.
	scan := &ScanResult{
		Files: []FileInfo{
			{Name: "main.go", Extension: "go", Size: 100},
			{Name: "app.ts", Extension: "ts", Size: 200},
			{Name: "readme.md", Extension: "md", Size: 500},         // not source
			{Name: "package.json", Extension: "json", Size: 100},     // not source
			{Name: "empty.go", Extension: "go", Size: 0},             // zero size
			{Name: "huge.go", Extension: "go", Size: 100 * 1024},     // too big
			{Name: "style.css", Extension: "css", Size: 100},         // not source
		},
	}

	files := selectSourceFiles(scan)

	nameSet := map[string]bool{}
	for _, f := range files {
		nameSet[f.Name] = true
	}

	if !nameSet["main.go"] {
		t.Error("expected main.go to be selected")
	}
	if !nameSet["app.ts"] {
		t.Error("expected app.ts to be selected")
	}
	if nameSet["readme.md"] {
		t.Error("readme.md should not be selected (not source)")
	}
	if nameSet["empty.go"] {
		t.Error("empty.go should not be selected (zero size)")
	}
	if nameSet["huge.go"] {
		t.Error("huge.go should not be selected (too big)")
	}
}

func TestDetectConventions_Integration(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "handler.go", `package pkg

import (
	"fmt"
	"os"
)

func HandleRequest() error {
	if err := doWork(); err != nil {
		return fmt.Errorf("handle: %w", err)
	}
	return nil
}

func doWork() error {
	_ = os.Getenv("HOME")
	return nil
}
`)
	writeFile(t, root, "handler_test.go", `package pkg

import "testing"

func TestHandleRequest(t *testing.T) {}
`)

	scan, err := Scan(root)
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}

	conventions := DetectConventions(root, scan)

	// We should get at least some conventions from the Go file.
	if len(conventions) == 0 {
		t.Fatal("expected at least one convention from Go source")
	}

	categories := map[string]bool{}
	for _, c := range conventions {
		categories[c.Category] = true
	}

	if !categories["naming"] {
		t.Error("expected naming convention")
	}
}
