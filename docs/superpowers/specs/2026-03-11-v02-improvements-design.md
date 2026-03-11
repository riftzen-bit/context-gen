# context-gen v0.2 Improvements

**Date:** 2026-03-11
**Status:** Approved

## Summary

Three improvements to context-gen: (1) CLAUDE.md generation aligned with Anthropic's official guidelines, (2) modern CLI UI using charmbracelet ecosystem, (3) MIT license.

## Motivation

- Generated CLAUDE.md is too verbose and misses sections Anthropic recommends (Code Style, Workflow, Developer Setup). Anthropic advises keeping CLAUDE.md under 200 lines with only information Claude can't infer from code.
- CLI UI uses plain `fmt.Println` with basic ANSI codes. A professional-looking UI improves user experience and trust.
- No license file means unclear usage rights for distribution.

---

## Phase 1: CLAUDE.md Generator — Anthropic Compliance

### Problem

Current generator outputs long listings of tech stack details that Claude can already infer. Missing sections that require human input have no placeholder. Output exceeds 200-line recommendation.

### Design

#### New Output Structure

```markdown
# Project Name

Brief description (from README.md first paragraph, or module path).

## Tech Stack
Go 1.26 | Docker | GitHub Actions
(concise, one-line summary)

## Common Commands
\```bash
go build ./...
go test ./...
go vet ./...
\```

## Code Style
- Indent: tabs (detected from .editorconfig)
- Line length: 120 (detected from .editorconfig)
- Formatter: gofmt
<!-- Add project-specific style rules here -->

## Project Structure
\```
cmd/       — CLI entry points
analyzer/  — Codebase scanning & detection
generator/ — Output file generation
\```

## Workflow
<!-- Add git workflow, branch naming, PR conventions here -->

## Architecture
<!-- Add key design decisions here -->
```

#### Smart Config Detection

New `analyzer/config_reader.go` reads config files for additional context:

| Config file | Extracts |
|---|---|
| `.editorconfig` | indent_style, indent_size, max_line_length |
| `.prettierrc` / `.prettierrc.json` | printWidth, tabWidth, singleQuote, semi |
| `tsconfig.json` | strict, target, baseUrl, paths |
| `.eslintrc.json` | key rule overrides (only JSON format, skip JS/YAML variants) |
| `Makefile` | target names via regex: lines matching `^targetname:` |
| `justfile` | deferred to future version (custom syntax, low priority) |
| `README.md` | first paragraph as project description |
| `package.json` scripts | dev, build, test, lint, start commands |

#### Type Changes

`analyzer/types.go` adds:

```go
type CodeStyle struct {
    IndentStyle  string // "tabs" or "spaces"
    IndentSize   int
    LineLength   int
    Formatter    string // "gofmt", "prettier", "black", etc.
    ExtraRules   []string
}

type ProjectInfo struct {
    // ... existing fields ...
    Name        string            // from go.mod module, package.json name, or directory name
    CodeStyle   CodeStyle
    Scripts     map[string]string // "build" -> "go build ./...", etc.
    Description string            // from README.md
}
```

#### Project Name Detection

Priority order:
1. `package.json` → `name` field
2. `go.mod` → last segment of module path (e.g. `github.com/paul/context-gen` → `context-gen`)
3. `Cargo.toml` → `[package] name`
4. `pyproject.toml` → `[project] name`
5. Fallback: directory name

#### Formatter Detection

Default formatter per primary language:
- Go → `gofmt`
- Python → `black` (if `pyproject.toml` has `[tool.black]`), else `none`
- JS/TS → `prettier` (if `.prettierrc*` exists), else `none`
- Rust → `rustfmt`

Override: if a formatter config file exists (`.prettierrc`, `pyproject.toml [tool.black]`, `rustfmt.toml`), use that formatter regardless of language default.

#### Scripts vs Hardcoded Commands

Priority for Common Commands section:
1. **Detected scripts** from `package.json scripts` or `Makefile` targets take precedence — they reflect the project's actual workflow.
2. **Hardcoded language defaults** (current `writeBuildCommands()`) are used as fallback only when no scripts are detected.
3. If both exist, detected scripts win. The hardcoded commands are never mixed with detected scripts.

#### README Parsing Heuristic

To extract project description from README.md:
1. Skip lines starting with `#`, `!`, `[`, `<`, or badge patterns (`[![`)
2. Skip empty lines
3. Take the first non-empty text block (until next empty line or heading)
4. Truncate to 200 characters
5. If no suitable text found, fall back to module/package name

#### .editorconfig Section Priority

Read the `[*]` (root/wildcard) section for global settings. Do not attempt to parse per-file-type sections — keep it simple.

#### Directory Annotations

Directory descriptions in Project Structure section use a hardcoded map of common directory names:
- `cmd/` → "CLI entry points"
- `pkg/` → "Public library code"
- `internal/` → "Private application code"
- `api/` → "API definitions"
- `web/` → "Web assets"
- `src/` → "Source code"
- `lib/` → "Library code"
- `test/` / `tests/` → "Tests"
- `docs/` → "Documentation"

Unknown directories get no annotation.

#### Cursor Format (.cursorrules)

`.cursorrules` gets the same concise template overhaul as CLAUDE.md. The only differences:
- Header: `# Project Context` (instead of project name)
- No `## Workflow` or `## Architecture` placeholder sections (Cursor doesn't use these)
- Everything else (Tech Stack, Commands, Code Style, Structure) is identical

#### Backward Compatibility Note

The new template format is a **breaking change** from v0.1 output. Running `context-gen update` will overwrite existing CLAUDE.md with the new format. This is acceptable — the `update` command's purpose is to regenerate from scratch. Users who have customized their CLAUDE.md should use `init` which creates `.generated` files instead of overwriting.

#### Generator Changes

- Template rewritten to follow Anthropic guidelines
- Concise tech stack (one line, pipe-separated)
- Only include sections with detected data + placeholder sections for human input
- Target: under 200 lines output
- Description directories with purpose annotations (` — description`)

#### Files Changed

- **New:** `analyzer/config_reader.go` — config file parsing
- **New:** `analyzer/config_reader_test.go` — tests
- **Modified:** `analyzer/types.go` — add CodeStyle, Scripts, Description
- **Modified:** `analyzer/detector.go` — call config reader
- **Modified:** `generator/generator.go` — new template
- **Modified:** `generator/generator_test.go` — updated tests

---

## Phase 2: CLI UI with Charmbracelet

### Problem

Plain `fmt.Println` + manual ANSI codes produce a basic-looking CLI. No spinner during scanning, no arrow-key navigation, no styled boxes.

### Design

#### Dependencies

```
github.com/charmbracelet/lipgloss    — declarative terminal styling
github.com/charmbracelet/bubbletea   — Elm-architecture TUI framework
github.com/charmbracelet/bubbles     — ready-made components (spinner, list)
```

#### UI Components

**Welcome banner** — lipgloss bordered box:
```
╭──────────────────────────────────────╮
│  context-gen v0.1.0                  │
│  Generate AI context files           │
╰──────────────────────────────────────╯
```

**Menu** — arrow key navigation via bubbles/list:
```
  What would you like to do?

  > Generate context files
    Update existing files
    Preview output
    Help
    Exit
```

**Spinner** — during scan/detect:
```
  ⣾ Scanning /home/user/project...
```

**Results table** — lipgloss bordered:
```
╭─ Detection Results ─────────────────╮
│ Languages   Go (80%), Shell (20%)   │
│ Build       Go Modules, Make        │
│ Testing     go test                 │
│ Linting     golangci-lint           │
╰─────────────────────────────────────╯
```

**Success/error** — styled messages:
```
  ✓ Created CLAUDE.md
  ✓ Created .cursorrules
  ✗ Error: directory not found
```

#### Architecture

New `ui/` package:

- `ui/styles.go` — lipgloss style definitions (colors, borders, text)
- `ui/menu.go` — bubbletea Model for main menu (arrow keys + enter)
- `ui/spinner.go` — bubbletea Model for scanning spinner
- `ui/prompts.go` — styled text input and format selection
- `ui/results.go` — detection results display

Modified:
- `cmd/interactive.go` — replace fmt-based menu with bubbletea program
- `cmd/root.go` — use lipgloss for CLI mode output (non-interactive)

#### Backward Compatibility

- CLI flags (`context-gen init -d . -f claude`) work unchanged
- Non-interactive output uses lipgloss styling but no bubbletea
- Non-TTY detection: when stdout is piped, fallback to plain text (no colors, no spinner)
- All existing tests remain valid

#### TTY Detection

```go
func isTTY() bool {
    _, err := unix.IoctlGetTermios(int(os.Stdout.Fd()), unix.TCGETS)
    return err == nil
}
```

Note: `lipgloss.HasDarkBackground()` detects color scheme, NOT TTY status. Use the `isTTY()` function above for TTY detection. Use `lipgloss.HasDarkBackground()` separately if adaptive color theming is desired.

If bubbletea initialization fails at runtime (unsupported terminal), fall back to the current `fmt.Println` flow gracefully. Wrap `tea.NewProgram().Run()` in error handling that triggers the fallback path.

#### Terminal Width

UI components should adapt to terminal width. Use `lipgloss.Width()` and `os.Stdout` terminal size detection. Minimum supported width: 40 columns. Below that, boxes degrade to unbordered text.

---

## Phase 3: MIT License

### Files

**New: `LICENSE`**
Standard MIT License text with `Copyright (c) 2026 Paul`.

**Modified: `main.go`**
Add copyright header comment:
```go
// Copyright (c) 2026 Paul. Licensed under the MIT License.
// See LICENSE file in the project root for full license text.
```

**Modified: version command output (bump to v0.2.0)**
```
context-gen v0.2.0
Copyright (c) 2026 Paul
Licensed under the MIT License
```

Version string updated to v0.2.0 across: welcome banner, version command, and any other references.

Only `main.go` gets the header — no spam across all files.

---

## Testing Strategy

### Phase 1
- Unit tests for config_reader (parse .editorconfig, .prettierrc, package.json scripts, README)
- Updated generator tests validating new template structure
- Integration test: scan a fixture directory, verify output matches Anthropic format

### Phase 2
- UI component tests using bubbletea test helpers (simulated key events)
- Existing interactive_test.go adapted for new bubbletea models
- Manual testing for visual correctness

### Phase 3
- Verify LICENSE file exists and contains correct text
- Verify version command output includes copyright

## Execution Order

Phase 1 → Phase 2 → Phase 3. Each phase is independently valuable and testable.
