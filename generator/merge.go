package generator

import "strings"

// Section represents a named markdown section.
type Section struct {
	Name    string // e.g., "Tech Stack", "Workflow", "" for preamble
	Content string // full content including the ## header line
}

// ParseSections splits markdown content into ordered sections by "## " headers.
// Returns a slice of Section preserving order.
// Content before the first ## header is treated as section "" (preamble).
func ParseSections(content string) []Section {
	if content == "" {
		return nil
	}

	lines := strings.Split(content, "\n")
	var sections []Section
	var current strings.Builder
	currentName := ""
	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
		}

		if !inCodeBlock && strings.HasPrefix(line, "## ") && !strings.HasPrefix(line, "### ") {
			// Flush previous section
			if current.Len() > 0 || currentName != "" {
				sections = append(sections, Section{
					Name:    currentName,
					Content: current.String(),
				})
			}
			currentName = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			current.Reset()
			current.WriteString(line)
			current.WriteString("\n")
		} else {
			current.WriteString(line)
			current.WriteString("\n")
		}
	}

	// Flush last section
	if current.Len() > 0 || currentName != "" {
		sections = append(sections, Section{
			Name:    currentName,
			Content: current.String(),
		})
	}

	// Skip empty preamble (only whitespace)
	if len(sections) > 0 && sections[0].Name == "" && strings.TrimSpace(sections[0].Content) == "" {
		sections = sections[1:]
	}

	return sections
}

// MergeSections merges generated content with existing content.
// Rules:
//  1. If existing section still has placeholder (<!-- ... -->), replace with generated
//  2. If existing section content is identical to generated, replace with new generated
//  3. If existing section was customized by user, keep existing version
//  4. New sections from generated that don't exist in existing are appended
//  5. Sections only in existing (user-added) are preserved
//  6. Order follows the existing file's section order
func MergeSections(existing, generated string) string {
	if existing == "" {
		return generated
	}
	if generated == "" {
		return existing
	}

	existingSections := ParseSections(existing)
	generatedSections := ParseSections(generated)

	// Build lookup maps for generated sections by name
	genByName := make(map[string]string, len(generatedSections))
	for _, s := range generatedSections {
		genByName[s.Name] = s.Content
	}

	// Track which generated sections are used
	usedGen := make(map[string]bool)

	// Build result starting with existing order
	var result []Section
	for _, es := range existingSections {
		genContent, inGenerated := genByName[es.Name]
		if !inGenerated {
			// User-added section not in generated: preserve it
			result = append(result, es)
			continue
		}

		usedGen[es.Name] = true

		if isPlaceholder(es.Content) || es.Content == genContent {
			// Section is placeholder or identical to generated: use new generated
			result = append(result, Section{Name: es.Name, Content: genContent})
		} else {
			// User customized this section: keep existing
			result = append(result, es)
		}
	}

	// Append new sections from generated that aren't in existing
	for _, gs := range generatedSections {
		if !usedGen[gs.Name] {
			result = append(result, gs)
		}
	}

	// Build final output
	var b strings.Builder
	for _, s := range result {
		b.WriteString(s.Content)
	}

	return b.String()
}

// isPlaceholder returns true if the section content contains an HTML comment.
func isPlaceholder(content string) bool {
	return strings.Contains(content, "<!--")
}
