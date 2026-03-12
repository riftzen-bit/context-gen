package analyzer

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// GitIgnore holds parsed .gitignore patterns for matching.
type GitIgnore struct {
	patterns []ignorePattern
}

// ignorePattern represents a single .gitignore rule.
type ignorePattern struct {
	pattern  string
	negate   bool
	dirOnly  bool
	anchored bool // pattern contains / (not trailing), anchored to root
}

// ParseGitIgnore reads a .gitignore file and returns a GitIgnore matcher.
// Returns an empty matcher (matches nothing) if the file doesn't exist.
func ParseGitIgnore(root string) *GitIgnore {
	gi := &GitIgnore{}

	f, err := os.Open(filepath.Join(root, ".gitignore"))
	if err != nil {
		return gi
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		p := parseLine(line)
		if p != nil {
			gi.patterns = append(gi.patterns, *p)
		}
	}

	return gi
}

// parseLine parses a single .gitignore line into a pattern.
// Returns nil for comments and empty lines.
func parseLine(line string) *ignorePattern {
	// Trim trailing whitespace
	line = strings.TrimRight(line, " \t\r")
	// Trim leading whitespace for pattern matching
	line = strings.TrimLeft(line, " \t")

	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	p := &ignorePattern{}

	// Negation
	if strings.HasPrefix(line, "!") {
		p.negate = true
		line = line[1:]
	}

	// Remove leading/trailing slashes, but remember them
	if strings.HasSuffix(line, "/") {
		p.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}

	// If pattern contains a slash (not trailing), it's anchored to root
	if strings.Contains(line, "/") {
		p.anchored = true
		line = strings.TrimPrefix(line, "/")
	}

	p.pattern = line
	return p
}

// IsIgnored returns true if the given relative path should be ignored.
// isDir indicates whether the path is a directory.
func (gi *GitIgnore) IsIgnored(relPath string, isDir bool) bool {
	if len(gi.patterns) == 0 {
		return false
	}

	// Normalize path separators to forward slashes for matching
	normalized := filepath.ToSlash(relPath)

	ignored := false
	for _, p := range gi.patterns {
		if p.dirOnly && !isDir {
			continue
		}

		if matchPattern(p, normalized) {
			ignored = !p.negate
		}
	}

	return ignored
}

// matchPattern checks if a normalized path matches an ignore pattern.
func matchPattern(p ignorePattern, path string) bool {
	pattern := p.pattern

	if p.anchored {
		// Anchored: match against the full relative path
		return matchGlob(pattern, path)
	}

	// Unanchored: match against any path segment or suffix
	// e.g., "*.log" matches "foo/bar/debug.log"
	// e.g., "build" matches "build" and "sub/build"

	// Try matching against the full path
	if matchGlob(pattern, path) {
		return true
	}

	// Try matching against each path suffix
	parts := strings.Split(path, "/")
	for i := range parts {
		suffix := strings.Join(parts[i:], "/")
		if matchGlob(pattern, suffix) {
			return true
		}
	}

	return false
}

// matchGlob performs glob-style matching with support for:
// - * matches anything except /
// - ** matches everything including /
// - ? matches any single character except /
func matchGlob(pattern, str string) bool {
	// Handle ** (double star) by expanding
	if strings.Contains(pattern, "**") {
		return matchDoubleStar(pattern, str)
	}

	return matchSimpleGlob(pattern, str)
}

// matchSimpleGlob handles *, ?, and character matching (no **).
func matchSimpleGlob(pattern, str string) bool {
	px, sx := 0, 0
	starPx, starSx := -1, -1

	for sx < len(str) {
		if px < len(pattern) && (pattern[px] == '?' && str[sx] != '/') {
			px++
			sx++
		} else if px < len(pattern) && pattern[px] == '*' {
			starPx = px
			starSx = sx
			px++
		} else if px < len(pattern) && pattern[px] == str[sx] {
			px++
			sx++
		} else if starPx >= 0 {
			// Backtrack: advance star match by one (but not past /)
			starSx++
			if starSx <= len(str) && (starSx == len(str) || str[starSx-1] != '/') {
				sx = starSx
				px = starPx + 1
			} else {
				return false
			}
		} else {
			return false
		}
	}

	// Consume remaining *'s in pattern
	for px < len(pattern) && pattern[px] == '*' {
		px++
	}

	return px == len(pattern)
}

// matchDoubleStar handles ** patterns by splitting on ** and matching segments.
func matchDoubleStar(pattern, str string) bool {
	parts := strings.SplitN(pattern, "**", 2)
	prefix := parts[0]
	suffix := parts[1]

	// Remove leading / from suffix (** absorbs it)
	suffix = strings.TrimPrefix(suffix, "/")

	// If prefix is empty or matches the start of str
	if prefix != "" {
		prefix = strings.TrimSuffix(prefix, "/")
		if !strings.HasPrefix(str, prefix) {
			return false
		}
		// Advance past the prefix
		str = strings.TrimPrefix(str, prefix)
		str = strings.TrimPrefix(str, "/")
	}

	if suffix == "" {
		return true // ** at end matches everything
	}

	// Try matching suffix against every possible position
	for i := 0; i <= len(str); i++ {
		candidate := str[i:]
		if matchGlob(suffix, candidate) {
			return true
		}
		// Only advance to next / boundary or single character
		if i < len(str) && str[i] == '/' {
			continue
		}
	}

	return false
}

// HasPatterns returns true if any patterns were loaded.
func (gi *GitIgnore) HasPatterns() bool {
	return len(gi.patterns) > 0
}
