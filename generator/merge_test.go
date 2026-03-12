package generator

import (
	"strings"
	"testing"
)

func TestParseSections_Basic(t *testing.T) {
	input := "# My Project\n\nSome intro text.\n\n## Tech Stack\n\nGo | Docker\n\n## Commands\n\n```bash\ngo build\n```\n\n## Workflow\n\n<!-- Add workflow here -->\n\n"

	sections := ParseSections(input)

	if len(sections) != 4 {
		t.Fatalf("expected 4 sections (preamble + 3), got %d", len(sections))
	}

	// Preamble
	if sections[0].Name != "" {
		t.Errorf("expected preamble name to be empty, got %q", sections[0].Name)
	}
	if !strings.Contains(sections[0].Content, "# My Project") {
		t.Error("preamble should contain the top-level heading")
	}

	// Tech Stack
	if sections[1].Name != "Tech Stack" {
		t.Errorf("expected section 1 name 'Tech Stack', got %q", sections[1].Name)
	}
	if !strings.Contains(sections[1].Content, "Go | Docker") {
		t.Error("Tech Stack section should contain tech content")
	}

	// Commands
	if sections[2].Name != "Commands" {
		t.Errorf("expected section 2 name 'Commands', got %q", sections[2].Name)
	}

	// Workflow
	if sections[3].Name != "Workflow" {
		t.Errorf("expected section 3 name 'Workflow', got %q", sections[3].Name)
	}
	if !strings.Contains(sections[3].Content, "<!--") {
		t.Error("Workflow section should contain placeholder comment")
	}
}

func TestParseSections_CodeBlockWithHash(t *testing.T) {
	input := "## Commands\n\n```bash\n## This is a bash comment, not a section\ngo build\n```\n\n## Real Section\n\nContent here.\n"

	sections := ParseSections(input)

	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}
	if sections[0].Name != "Commands" {
		t.Errorf("expected 'Commands', got %q", sections[0].Name)
	}
	if sections[1].Name != "Real Section" {
		t.Errorf("expected 'Real Section', got %q", sections[1].Name)
	}
	// The code block content should be inside Commands, not split
	if !strings.Contains(sections[0].Content, "## This is a bash comment") {
		t.Error("code block with ## should stay within Commands section")
	}
}

func TestParseSections_NoPreamble(t *testing.T) {
	input := "## First Section\n\nContent.\n\n## Second Section\n\nMore content.\n"

	sections := ParseSections(input)

	if len(sections) != 2 {
		t.Fatalf("expected 2 sections (no preamble), got %d", len(sections))
	}
	if sections[0].Name != "First Section" {
		t.Errorf("expected 'First Section', got %q", sections[0].Name)
	}
	if sections[1].Name != "Second Section" {
		t.Errorf("expected 'Second Section', got %q", sections[1].Name)
	}
}

func TestMergeSections_AllPlaceholders(t *testing.T) {
	existing := "# Project\n\n## Tech Stack\n\n<!-- Add tech stack -->\n\n## Workflow\n\n<!-- Add workflow -->\n\n"
	generated := "# Project\n\n## Tech Stack\n\nGo | Docker\n\n## Workflow\n\ngit flow\n\n"

	result := MergeSections(existing, generated)

	if !strings.Contains(result, "Go | Docker") {
		t.Error("placeholder Tech Stack should be replaced with generated content")
	}
	if !strings.Contains(result, "git flow") {
		t.Error("placeholder Workflow should be replaced with generated content")
	}
	if strings.Contains(result, "<!-- Add tech stack -->") {
		t.Error("placeholder comment should be removed")
	}
	if strings.Contains(result, "<!-- Add workflow -->") {
		t.Error("placeholder comment should be removed")
	}
}

func TestMergeSections_UserEditedWorkflow(t *testing.T) {
	existing := "# Project\n\n## Tech Stack\n\n<!-- Add tech stack -->\n\n## Workflow\n\nWe use trunk-based development.\n\n"
	generated := "# Project\n\n## Tech Stack\n\nGo | Docker\n\n## Workflow\n\n<!-- Add git workflow here -->\n\n"

	result := MergeSections(existing, generated)

	// Tech Stack was placeholder, should be replaced
	if !strings.Contains(result, "Go | Docker") {
		t.Error("placeholder Tech Stack should be replaced")
	}
	// Workflow was user-edited, should be preserved
	if !strings.Contains(result, "We use trunk-based development.") {
		t.Error("user-edited Workflow should be preserved")
	}
	if strings.Contains(result, "<!-- Add git workflow here -->") {
		t.Error("generated placeholder should not replace user content")
	}
}

func TestMergeSections_UserEditedArchitecture(t *testing.T) {
	existing := "# Project\n\n## Tech Stack\n\n<!-- Add tech stack -->\n\n## Architecture\n\nClean architecture with hexagonal ports.\n\n"
	generated := "# Project\n\n## Tech Stack\n\nGo | Docker\n\n## Architecture\n\n<!-- Add key design decisions here -->\n\n"

	result := MergeSections(existing, generated)

	if !strings.Contains(result, "Go | Docker") {
		t.Error("placeholder Tech Stack should be replaced")
	}
	if !strings.Contains(result, "Clean architecture with hexagonal ports.") {
		t.Error("user-edited Architecture should be preserved")
	}
}

func TestMergeSections_UserAddedCustomSection(t *testing.T) {
	existing := "# Project\n\n## Tech Stack\n\n<!-- Add tech stack -->\n\n## Deployment\n\nDeploy via Kubernetes.\n\n"
	generated := "# Project\n\n## Tech Stack\n\nGo | Docker\n\n"

	result := MergeSections(existing, generated)

	if !strings.Contains(result, "Go | Docker") {
		t.Error("placeholder Tech Stack should be replaced")
	}
	if !strings.Contains(result, "## Deployment") {
		t.Error("user-added Deployment section should be preserved")
	}
	if !strings.Contains(result, "Deploy via Kubernetes.") {
		t.Error("Deployment content should be preserved")
	}
}

func TestMergeSections_NewSectionInGenerated(t *testing.T) {
	existing := "# Project\n\n## Tech Stack\n\nGo | Docker\n\n"
	generated := "# Project\n\n## Tech Stack\n\nGo | Docker\n\n## Coding Guidelines\n\n- Use gofmt\n\n"

	result := MergeSections(existing, generated)

	if !strings.Contains(result, "## Coding Guidelines") {
		t.Error("new section from generated should be appended")
	}
	if !strings.Contains(result, "- Use gofmt") {
		t.Error("new section content should be present")
	}
}

func TestMergeSections_PreserveOrder(t *testing.T) {
	// Existing has Workflow before Tech Stack
	existing := "# Project\n\n## Workflow\n\n<!-- Add workflow -->\n\n## Tech Stack\n\n<!-- Add tech stack -->\n\n"
	generated := "# Project\n\n## Tech Stack\n\nGo | Docker\n\n## Workflow\n\ngit flow\n\n"

	result := MergeSections(existing, generated)

	workflowIdx := strings.Index(result, "## Workflow")
	techIdx := strings.Index(result, "## Tech Stack")

	if workflowIdx == -1 || techIdx == -1 {
		t.Fatal("both sections should be present")
	}
	if workflowIdx > techIdx {
		t.Error("existing order should be preserved: Workflow before Tech Stack")
	}
}

func TestMergeSections_EmptyExisting(t *testing.T) {
	generated := "# Project\n\n## Tech Stack\n\nGo | Docker\n\n"

	result := MergeSections("", generated)

	if result != generated {
		t.Errorf("empty existing should return generated as-is, got:\n%s", result)
	}
}

func TestMergeSections_EmptyGenerated(t *testing.T) {
	existing := "# Project\n\n## Tech Stack\n\nGo | Docker\n\n"

	result := MergeSections(existing, "")

	if result != existing {
		t.Errorf("empty generated should return existing as-is, got:\n%s", result)
	}
}
