package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/riftzen-bit/context-gen/generator"
)

func TestDetectExistingFormats_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	formats := detectExistingFormats(dir)
	if len(formats) != 0 {
		t.Errorf("expected empty slice, got %v", formats)
	}
}

func TestDetectExistingFormats_OnlyClaude(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# test"), 0644)

	formats := detectExistingFormats(dir)
	if len(formats) != 1 {
		t.Fatalf("expected 1 format, got %d: %v", len(formats), formats)
	}
	if formats[0] != generator.FormatClaude {
		t.Errorf("expected FormatClaude, got %s", formats[0])
	}
}

func TestDetectExistingFormats_ClaudeAndCursor(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# test"), 0644)
	os.WriteFile(filepath.Join(dir, ".cursorrules"), []byte("# test"), 0644)

	formats := detectExistingFormats(dir)
	if len(formats) != 2 {
		t.Fatalf("expected 2 formats, got %d: %v", len(formats), formats)
	}
	if formats[0] != generator.FormatClaude {
		t.Errorf("expected formats[0] = FormatClaude, got %s", formats[0])
	}
	if formats[1] != generator.FormatCursor {
		t.Errorf("expected formats[1] = FormatCursor, got %s", formats[1])
	}
}

func TestDetectExistingFormats_ClineAndWindsurf_NotAll(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".clinerules"), []byte("# test"), 0644)
	os.WriteFile(filepath.Join(dir, ".windsurfrules"), []byte("# test"), 0644)

	formats := detectExistingFormats(dir)
	if len(formats) != 2 {
		t.Fatalf("expected 2 formats, got %d: %v", len(formats), formats)
	}
	if formats[0] != generator.FormatCline {
		t.Errorf("expected formats[0] = FormatCline, got %s", formats[0])
	}
	if formats[1] != generator.FormatWindsurf {
		t.Errorf("expected formats[1] = FormatWindsurf, got %s", formats[1])
	}
	// Verify FormatAll is NOT returned (the bug this fixes)
	for _, f := range formats {
		if f == generator.FormatAll {
			t.Errorf("should not return FormatAll, got %v", formats)
		}
	}
}
