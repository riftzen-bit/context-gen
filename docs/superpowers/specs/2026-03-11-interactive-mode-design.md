# Interactive Mode Design

**Date:** 2026-03-11
**Status:** Approved

## Summary

Add a menu-driven interactive mode to context-gen. When the user runs `context-gen` without arguments, they enter an interactive menu instead of seeing help text. All existing subcommands (`init`, `update`, `preview`) continue to work as before.

## Motivation

Users forget CLI commands and flags. An interactive menu lets them discover and use features without memorizing syntax.

## Design

### Trigger

`len(os.Args) == 1` (no arguments) → `runInteractive()` instead of `printUsage()`.

### Main Menu

```
context-gen v0.1.0
Generate AI context files for your codebase

What would you like to do?

1. Generate context files (first time)
2. Update existing context files
3. Preview output
4. Help
5. Exit

>
```

### Sub-prompts

Each menu item prompts for required inputs:

**Generate (init):**
1. Target directory (default: `.`)
2. Format selection: Claude / Cursor / Both
3. Output directory (default: same as target)

**Update:**
1. Target directory (default: `.`)
2. Format override (optional, auto-detects by default)

**Preview:**
1. Target directory (default: `.`)
2. Format selection

### After Action

After completing any action, print "Press Enter to continue..." and return to the main menu. User can exit via option 5 or Ctrl+C.

### Architecture

New file: `cmd/interactive.go`

Functions:
- `runInteractive()` — main loop: display menu, read choice, dispatch, repeat
- `promptChoice(prompt string, options []string) int` — numbered list, reads integer, validates range, retries on invalid input
- `promptString(prompt, defaultVal string) string` — text input with default value shown in brackets
- `promptConfirm(prompt string) bool` — y/n prompt

Interactive mode reuses existing `runInit`, `runUpdate`, `runPreview` by constructing the appropriate `[]string` args and calling them directly. No refactoring of existing code needed.

### Error Handling

- Invalid menu choice → print error, re-prompt (no crash)
- Invalid directory → print error, re-prompt
- Ctrl+C / EOF → clean exit
- All prompts use `bufio.Reader` for line-based input

### Dependencies

None. Pure `fmt`, `bufio`, `os`, `strings` from stdlib.

### Backward Compatibility

Fully backward compatible. All existing CLI commands and flags work unchanged. Interactive mode is purely additive.

## Testing

- Unit tests for `promptChoice`, `promptString` by injecting a custom `io.Reader`
- Integration test: verify `runInteractive` dispatches correctly based on input sequence
