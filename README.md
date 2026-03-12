<p align="center">
  <h1 align="center">context-gen</h1>
  <p align="center">
    <strong>One scan. Every AI assistant. Zero config.</strong>
  </p>
  <p align="center">
    <a href="#installation">Install</a> · <a href="#quick-start">Quick Start</a> · <a href="#supported-formats">Formats</a> · <a href="#how-it-works">How It Works</a>
  </p>
</p>

---

**context-gen** scans your project, detects languages, frameworks, build tools, code style, and conventions — then generates context files that AI coding assistants use to understand your codebase.

No API keys. No LLM calls. Runs locally, fast, and deterministic.

```
$ context-gen init -f all

 Scanning /home/user/my-app
  Found 142 files in 18 directories

 Detection Results
  Languages:  TypeScript, CSS
  Frameworks: React, Tailwind CSS
  Build:      npm (vite)
  CI/CD:      GitHub Actions
  Style:      prettier, 2-space indent

 Writing files
  CREATE CLAUDE.md
  CREATE .cursorrules
  CREATE AGENTS.md
  CREATE .cursor/rules/project.mdc
  CREATE .clinerules
  CREATE .windsurfrules

 Done! Review the generated files and customize them for your project.
```

## Why context-gen?

AI coding assistants work better when they understand your project. But writing and maintaining context files by hand is tedious — especially across multiple tools.

**context-gen automates this.** One command gives every AI assistant the same, accurate understanding of your codebase:

- **Accurate detection** — reads your actual configs, not guesses
- **Framework-aware rules** — generates coding guidelines specific to your stack (React hooks, Go error handling, Rust Result types...)
- **Smart updates** — re-scan without losing your customizations
- **6 formats** — covers every major AI coding assistant

## Supported Formats

| Format | File | AI Assistant |
|--------|------|--------------|
| Claude | `CLAUDE.md` | [Claude Code](https://docs.anthropic.com/en/docs/claude-code) |
| Cursor | `.cursorrules` | [Cursor](https://cursor.com) |
| Cursor MDC | `.cursor/rules/project.mdc` | Cursor (new rules format) |
| Copilot | `AGENTS.md` | [GitHub Copilot](https://github.com/features/copilot) |
| Cline | `.clinerules` | [Cline](https://github.com/cline/cline) / Roo Code |
| Windsurf | `.windsurfrules` | [Windsurf](https://windsurf.com) |

Generate one format, a pair (`both` = Claude + Cursor), or all six at once.

## Installation

### Go Install

```bash
go install github.com/riftzen-bit/context-gen@latest
```

### From Source

```bash
git clone https://github.com/riftzen-bit/context-gen.git
cd context-gen
go build -o context-gen .
```

### Pre-built Binaries

Download from [Releases](https://github.com/riftzen-bit/context-gen/releases).

## Quick Start

```bash
# Interactive mode — menu-driven, no flags needed
context-gen

# Generate CLAUDE.md + .cursorrules (default)
context-gen init

# Generate for a specific format
context-gen init -f claude

# Generate ALL formats at once
context-gen init -f all

# Scan a different directory, output elsewhere
context-gen init -d ./my-project -o ./output

# Update existing files (preserves your edits)
context-gen update

# Preview what would be generated
context-gen preview
```

## Commands

| Command | Description |
|---------|-------------|
| `context-gen` | Interactive mode with arrow-key navigation |
| `context-gen init` | Scan project and generate context files |
| `context-gen update` | Re-scan and update (smart merge preserves edits) |
| `context-gen preview` | Preview generated output without writing |
| `context-gen version` | Show version info |
| `context-gen help` | Show usage help |

### Flags

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--dir` | `-d` | Target directory to scan | `.` |
| `--format` | `-f` | Output format | `both` |
| `--output` | `-o` | Output directory | scanned dir |
| `--dry-run` | | Preview without writing (init) | |
| `--force` | | Skip smart merge, overwrite all (update) | |

**Format options:** `claude`, `cursor`, `agents`, `cursor-mdc`, `cline`, `windsurf`, `both`, `all`

## What Gets Detected

context-gen reads your actual project files — not just file extensions.

| Category | Sources | Examples |
|----------|---------|----------|
| **Languages** | File extensions, configs | Go, TypeScript, JavaScript, Python, Rust, Java, Kotlin, C#, C++, Ruby, PHP, Swift, Dart, Elixir, and more |
| **Frameworks** | `package.json`, `Cargo.toml`, `go.mod`, imports | React, Next.js, Vue, Svelte, Angular, Tauri, Electron, Express, FastAPI, Django, Rails, and more |
| **Build tools** | Lock files, configs | npm/yarn/pnpm/bun, Cargo, Go Modules, Maven, Gradle, Make, CMake |
| **CI/CD** | Config directories | GitHub Actions, GitLab CI, CircleCI, Travis CI, Jenkins |
| **Code style** | `.editorconfig`, `.prettierrc`, `pyproject.toml` | Indentation, formatters, linters |
| **Conventions** | Source code analysis | Naming patterns, import styles, error handling |
| **Scripts** | `package.json`, `Makefile` | dev, build, test, lint commands |

## Smart Update

`context-gen update` preserves your work:

```
Sections with placeholder comments  →  Updated with fresh scan data
Sections you've customized          →  Kept exactly as-is
New sections from the scanner       →  Appended at the end
Custom sections you added manually  →  Preserved
```

It also auto-detects which formats you're using — no need to specify `--format` on updates.

Use `--force` to overwrite everything and start fresh.

## Framework-Aware Guidelines

Generated context files include coding guidelines tailored to your detected stack:

- **React** → functional components, hooks, memoization, composition
- **Go** → MixedCaps naming, error wrapping with `%w`, interface design
- **Rust** → `Result<T, E>`, no `unwrap()` in production, serde patterns
- **TypeScript** → strict mode, no `any`, type guards, `as const`
- **Next.js** → App Router, Server Components, route handlers
- **Tailwind CSS** → utility-first, avoid custom CSS, responsive design
- **Tauri** → Rust backend commands, IPC patterns, security
- And 20+ more languages and frameworks...

## Example Output

Running `context-gen init -f claude` on a React + TypeScript project:

````markdown
# my-app

A modern web application built with React and TypeScript.

## Tech Stack

TypeScript | React | Tailwind CSS | GitHub Actions

## Common Commands

```bash
pnpm install    # install dependencies
pnpm run dev    # start dev server
pnpm run build  # production build
pnpm run test   # run tests
pnpm run lint   # lint code
```

## Code Style

- Indent: space (2)
- Formatter: prettier
- Linters: ESLint

## Conventions

- **naming**: Files use PascalCase naming
- **imports**: Relative imports preferred over absolute
- **error_handling**: async/await with try/catch

## Coding Guidelines

- Use functional components with hooks, no class components
- Use strict TypeScript — no `any` types
- Use Tailwind utility classes, avoid custom CSS
- Memoize expensive renders with React.memo, useMemo, useCallback

## Testing

```bash
pnpm run test                          # Run all tests
pnpm run test -- path/to/file.test.ts  # Run single file
```
````

## How It Works

```
┌─────────┐     ┌──────────┐     ┌──────────┐
│  Scan   │────▶│  Detect  │────▶│ Generate │
│         │     │          │     │          │
│ Walk    │     │ Languages│     │ CLAUDE.md│
│ dirs    │     │ Frameworks│    │ .cursor  │
│ Collect │     │ Build    │     │ AGENTS.md│
│ metadata│     │ Style    │     │ .cline   │
│ .git-   │     │ CI/CD    │     │ .wind-   │
│  ignore │     │ Scripts  │     │  surf    │
└─────────┘     └──────────┘     └──────────┘
```

1. **Scan** — walks your project directory, respects `.gitignore`, collects file metadata
2. **Detect** — analyzes configs (`package.json`, `go.mod`, `Cargo.toml`, etc.), source patterns, and conventions
3. **Generate** — produces context files with framework-specific coding guidelines

Everything runs locally. No network calls. No API keys. Fast and deterministic.

## Interactive Mode

Just run `context-gen` with no arguments for a menu-driven experience:

```
╭──────────────────────────────╮
│ context-gen v0.4.0           │
│ Generate AI context files    │
╰──────────────────────────────╯

  Choose an action:
> Generate context files (init)
  Update existing files
  Preview output
  Quit

  Choose format:
> Both (CLAUDE.md + .cursorrules)
  Claude (CLAUDE.md)
  Cursor (.cursorrules)
  All formats
```

## Project Structure

```
context-gen/
├── main.go              # Entry point
├── cmd/
│   ├── root.go          # CLI command routing and flags
│   └── interactive.go   # TUI menu (bubbletea)
├── analyzer/
│   ├── scanner.go       # File system walker
│   ├── detector.go      # Language/framework detection
│   ├── conventions.go   # Code convention analysis
│   ├── config_reader.go # Config file parsing
│   └── gitignore.go     # .gitignore pattern matching
├── generator/
│   ├── generator.go     # Context file generation
│   ├── rules.go         # Framework-specific guidelines
│   └── merge.go         # Smart section merging
└── ui/
    ├── styles.go        # Terminal colors and styles
    ├── results.go       # Detection result formatting
    └── prompts.go       # User prompts
```

## Contributing

Contributions are welcome! Please open an issue first to discuss what you'd like to change.

```bash
# Clone and build
git clone https://github.com/riftzen-bit/context-gen.git
cd context-gen
go build ./...

# Run tests
go test ./... -v

# Run locally
go run . init -d /path/to/project
```

## License

[MIT](LICENSE) — use it however you want.
