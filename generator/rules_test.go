package generator

import (
	"strings"
	"testing"

	"github.com/riftzen-bit/context-gen/analyzer"
)

func TestRelevantRules_TauriReactProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "TypeScript", FileCount: 30, Percentage: 75},
			{Name: "Rust", FileCount: 10, Percentage: 25},
		},
		Frameworks: []string{"Tauri", "React", "Tailwind CSS"},
	}

	rules := RelevantRules(info)

	mustContain := []string{
		"Use strict TypeScript",       // TypeScript
		"Use Result<T, E>",            // Rust
		"#[tauri::command]",           // Tauri
		"functional components",       // React
		"Tailwind utility classes",    // Tailwind
		"No hardcoded secrets",        // General
	}
	for _, want := range mustContain {
		if !rulesContain(rules, want) {
			t.Errorf("expected rule containing %q", want)
		}
	}

	mustNotContain := []string{
		"$state, $derived",          // Svelte
		"snake_case for functions",  // Python
		"MixedCaps",                 // Go
		"Composition API",           // Vue
		"standalone components",     // Angular
		"App Router",                // Next.js
	}
	for _, unwanted := range mustNotContain {
		if rulesContain(rules, unwanted) {
			t.Errorf("unexpected rule containing %q", unwanted)
		}
	}
}

func TestRelevantRules_GoOnlyProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Go", FileCount: 20, Percentage: 100},
		},
	}

	rules := RelevantRules(info)

	if !rulesContain(rules, "MixedCaps") {
		t.Error("expected Go naming rule")
	}
	if !rulesContain(rules, "context.Context") {
		t.Error("expected Go context rule")
	}
	if !rulesContain(rules, "No hardcoded secrets") {
		t.Error("expected general rules")
	}

	mustNotContain := []string{
		"Result<T, E>",            // Rust
		"functional components",   // React
		"strict TypeScript",       // TypeScript
		"snake_case for functions", // Python
		"Composition API",         // Vue
	}
	for _, unwanted := range mustNotContain {
		if rulesContain(rules, unwanted) {
			t.Errorf("unexpected rule containing %q", unwanted)
		}
	}
}

func TestRelevantRules_EmptyProject(t *testing.T) {
	info := &analyzer.ProjectInfo{}

	rules := RelevantRules(info)

	if len(rules) != len(generalRules) {
		t.Errorf("expected %d general rules, got %d", len(generalRules), len(rules))
	}
	if !rulesContain(rules, "No hardcoded secrets") {
		t.Error("expected general rules")
	}
}

func TestRelevantRules_NextJSProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "TypeScript", FileCount: 50, Percentage: 100},
		},
		Frameworks: []string{"Next.js", "React", "Tailwind CSS"},
	}

	rules := RelevantRules(info)

	mustContain := []string{
		"strict TypeScript",       // TypeScript
		"App Router",              // Next.js
		"functional components",   // React
		"Tailwind utility classes", // Tailwind
		"No hardcoded secrets",    // General
	}
	for _, want := range mustContain {
		if !rulesContain(rules, want) {
			t.Errorf("expected rule containing %q", want)
		}
	}

	mustNotContain := []string{
		"MixedCaps",             // Go
		"Result<T, E>",          // Rust
		"Composition API",       // Vue
		"standalone components", // Angular
		"snake_case",            // Python
	}
	for _, unwanted := range mustNotContain {
		if rulesContain(rules, unwanted) {
			t.Errorf("unexpected rule containing %q", unwanted)
		}
	}
}

func TestRelevantRules_NoDuplicates(t *testing.T) {
	// Express, Fastify, and Hono share identical rules — verify no duplicates
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "TypeScript", FileCount: 20, Percentage: 100},
		},
		Frameworks: []string{"Express", "Fastify", "Hono"},
	}

	rules := RelevantRules(info)

	seen := make(map[string]bool)
	for _, r := range rules {
		if seen[r] {
			t.Errorf("duplicate rule: %q", r)
		}
		seen[r] = true
	}
}

func TestRelevantRules_JavaProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Java", FileCount: 30, Percentage: 100},
		},
	}

	rules := RelevantRules(info)

	if !rulesContain(rules, "records for data") {
		t.Error("expected Java records rule")
	}
	if !rulesContain(rules, "Optional") {
		t.Error("expected Java Optional rule")
	}
	if rulesContain(rules, "MixedCaps") {
		t.Error("unexpected Go rule")
	}
}

func TestRelevantRules_DjangoProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Python", FileCount: 40, Percentage: 100},
		},
		Frameworks: []string{"Django"},
	}

	rules := RelevantRules(info)

	if !rulesContain(rules, "type hints") {
		t.Error("expected Python type hints rule")
	}
	if !rulesContain(rules, "class-based views") {
		t.Error("expected Django class-based views rule")
	}
	if !rulesContain(rules, "Django REST Framework") {
		t.Error("expected Django REST Framework rule")
	}
	if rulesContain(rules, "strict TypeScript") {
		t.Error("unexpected TypeScript rule")
	}
}

func TestRelevantRules_RubyRailsProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Ruby", FileCount: 50, Percentage: 100},
		},
	}

	rules := RelevantRules(info)

	if !rulesContain(rules, "RuboCop") {
		t.Error("expected Ruby RuboCop rule")
	}
	if !rulesContain(rules, "symbols over strings") {
		t.Error("expected Ruby symbols rule")
	}
}

func TestRelevantRules_SwiftProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Swift", FileCount: 20, Percentage: 100},
		},
	}

	rules := RelevantRules(info)

	if !rulesContain(rules, "structs over classes") {
		t.Error("expected Swift structs rule")
	}
	if !rulesContain(rules, "guard for early returns") {
		t.Error("expected Swift guard rule")
	}
}

func TestRelevantRules_FastAPIProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "Python", FileCount: 30, Percentage: 100},
		},
		Frameworks: []string{"FastAPI"},
	}

	rules := RelevantRules(info)

	if !rulesContain(rules, "Pydantic models for request") {
		t.Error("expected FastAPI Pydantic rule")
	}
	if !rulesContain(rules, "dependency injection") {
		t.Error("expected FastAPI DI rule")
	}
}

func TestRelevantRules_ElectronProject(t *testing.T) {
	info := &analyzer.ProjectInfo{
		Languages: []analyzer.Language{
			{Name: "TypeScript", FileCount: 20, Percentage: 100},
		},
		Frameworks: []string{"Electron", "React"},
	}

	rules := RelevantRules(info)

	if !rulesContain(rules, "main process") {
		t.Error("expected Electron process separation rule")
	}
	if !rulesContain(rules, "contextBridge") {
		t.Error("expected Electron contextBridge rule")
	}
	if !rulesContain(rules, "functional components") {
		t.Error("expected React rule")
	}
}

// rulesContain checks if any rule contains the given substring.
func rulesContain(rules []string, sub string) bool {
	for _, r := range rules {
		if strings.Contains(r, sub) {
			return true
		}
	}
	return false
}
