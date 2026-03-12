package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/riftzen-bit/context-gen/analyzer"
	"github.com/riftzen-bit/context-gen/generator"
	"github.com/riftzen-bit/context-gen/ui"
)

func Execute() {
	args := os.Args[1:]

	if len(args) == 0 {
		runInteractive()
		return
	}

	if args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		printUsage()
		return
	}

	var err error
	switch args[0] {
	case "init":
		err = runInit(args[1:])
	case "update":
		err = runUpdate(args[1:])
	case "preview":
		err = runPreview(args[1:])
	case "version":
		fmt.Println("context-gen " + ui.Version)
		fmt.Println("Copyright (c) 2026 Paul")
		fmt.Println("Licensed under the MIT License")
	default:
		fmt.Println(ui.FormatError(fmt.Sprintf("Unknown command: %s", args[0])))
		fmt.Println()
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(ui.FormatError(err.Error()))
		os.Exit(1)
	}
}

// parseCommonFlags parses flags shared by init, update, and preview commands.
func parseCommonFlags(args []string) (targetDir string, format generator.Format, err error) {
	targetDir = "."
	format = generator.FormatBoth

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dir", "-d":
			if i+1 < len(args) {
				targetDir = args[i+1]
				i++
			}
		case "--format", "-f":
			if i+1 < len(args) {
				switch args[i+1] {
				case "claude":
					format = generator.FormatClaude
				case "cursor":
					format = generator.FormatCursor
				case "both":
					format = generator.FormatBoth
				case "agents":
					format = generator.FormatAgents
				case "cursor-mdc":
					format = generator.FormatCursorMDC
				case "cline":
					format = generator.FormatCline
				case "windsurf":
					format = generator.FormatWindsurf
				case "all":
					format = generator.FormatAll
				default:
					return "", "", fmt.Errorf("unknown format: %s (use: claude, cursor, agents, cursor-mdc, cline, windsurf, both, all)", args[i+1])
				}
				i++
			}
		}
	}

	return targetDir, format, nil
}

// parseOutputFlag extracts the --output/-o flag from args.
func parseOutputFlag(args []string) string {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--output", "-o":
			if i+1 < len(args) {
				return args[i+1]
			}
		}
	}
	return ""
}

// hasDryRun checks if --dry-run is present in args.
func hasDryRun(args []string) bool {
	for _, a := range args {
		if a == "--dry-run" {
			return true
		}
	}
	return false
}

// hasForce checks if --force is present in args.
func hasForce(args []string) bool {
	for _, a := range args {
		if a == "--force" {
			return true
		}
	}
	return false
}

// scanAndDetect runs the scan + detect pipeline on absDir, printing progress.
func scanAndDetect(absDir string) (*analyzer.ProjectInfo, error) {
	fmt.Println()
	fmt.Printf("%s %s\n", ui.Bold.Render("Scanning"), absDir)

	scan, err := analyzer.Scan(absDir)
	if err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}
	fmt.Printf("  Found %d files in %d directories\n", scan.TotalFiles, scan.TotalDirs)

	info, err := analyzer.Detect(absDir, scan)
	if err != nil {
		return nil, fmt.Errorf("detection error: %w", err)
	}

	fmt.Println()
	fmt.Println(ui.Bold.Render("Detection Results"))
	fmt.Println(ui.FormatResults(info))

	return info, nil
}

// resolveDir resolves and validates a directory path.
func resolveDir(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolving path: %w", err)
	}
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return "", fmt.Errorf("directory not found: %s", absDir)
	}
	return absDir, nil
}

// resolveOutputDir resolves the output directory, defaulting to absDir.
func resolveOutputDir(outputDir, absDir string) (string, error) {
	if outputDir == "" {
		return absDir, nil
	}
	outDir, err := filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("resolving output path: %w", err)
	}
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		return "", fmt.Errorf("output directory not found: %s", outDir)
	}
	return outDir, nil
}

func runInit(args []string) error {
	targetDir, format, err := parseCommonFlags(args)
	if err != nil {
		return err
	}
	dryRun := hasDryRun(args)
	outputDir := parseOutputFlag(args)

	absDir, err := resolveDir(targetDir)
	if err != nil {
		return err
	}

	outDir, err := resolveOutputDir(outputDir, absDir)
	if err != nil {
		return err
	}

	info, err := scanAndDetect(absDir)
	if err != nil {
		return err
	}

	files := generator.Generate(info, format)

	if dryRun {
		printDryRun(files)
		return nil
	}

	// Write files (init refuses to overwrite)
	fmt.Println()
	fmt.Println(ui.Bold.Render("Writing files"))
	for name, content := range files {
		outPath := filepath.Join(outDir, name)

		// Create parent directories if needed (for .cursor/rules/project.mdc)
		if dir := filepath.Dir(outPath); dir != outDir {
			os.MkdirAll(dir, 0755)
		}

		if _, err := os.Stat(outPath); err == nil {
			fmt.Println(ui.FormatFileSkipped(name, fmt.Sprintf("already exists, writing to %s.generated", name)))
			outPath = outPath + ".generated"
		}

		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			fmt.Println(ui.FormatError(fmt.Sprintf("writing %s: %v", name, err)))
			continue
		}
		fmt.Println(ui.FormatFileCreated(outPath))
	}

	fmt.Println()
	fmt.Println(ui.Success.Render("Done!") + " Review the generated files and customize them for your project.")
	return nil
}

func runUpdate(args []string) error {
	targetDir, formatFlag, err := parseCommonFlags(args)
	if err != nil {
		return err
	}
	outputDir := parseOutputFlag(args)
	force := hasForce(args)

	absDir, err := resolveDir(targetDir)
	if err != nil {
		return err
	}

	outDir, err := resolveOutputDir(outputDir, absDir)
	if err != nil {
		return err
	}

	// Auto-detect formats from existing files if no explicit --format given
	var formats []generator.Format
	if hasFlag(args, "--format", "-f") {
		formats = []generator.Format{formatFlag}
	} else {
		formats = detectExistingFormats(outDir)
		if len(formats) == 0 {
			formats = []generator.Format{formatFlag}
			fmt.Println(ui.Warning.Render("No existing context files found. Generating with format: " + string(formatFlag)))
		} else {
			var names []string
			for _, f := range formats {
				names = append(names, string(f))
			}
			fmt.Printf("  Detected existing formats: %s\n", ui.Bold.Render(strings.Join(names, ", ")))
		}
	}

	info, err := scanAndDetect(absDir)
	if err != nil {
		return err
	}

	files := make(map[string]string)
	for _, f := range formats {
		for k, v := range generator.Generate(info, f) {
			files[k] = v
		}
	}

	// Write files (smart merge preserves user edits unless --force)
	fmt.Println()
	fmt.Println(ui.Bold.Render("Updating files"))
	for name, content := range files {
		outPath := filepath.Join(outDir, name)

		// Create parent directories if needed (for .cursor/rules/project.mdc)
		if dir := filepath.Dir(outPath); dir != outDir {
			os.MkdirAll(dir, 0755)
		}

		action := ui.FormatFileCreated
		if existingData, err := os.ReadFile(outPath); err == nil {
			action = ui.FormatFileUpdated
			if !force {
				content = generator.MergeSections(string(existingData), content)
			}
		}

		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			fmt.Println(ui.FormatError(fmt.Sprintf("writing %s: %v", name, err)))
			continue
		}
		fmt.Println(action(outPath))
	}

	// Clean up .generated files if the real file is now written
	for name := range files {
		genPath := filepath.Join(outDir, name+".generated")
		if _, err := os.Stat(genPath); err == nil {
			fmt.Println(ui.FormatFileSkipped(name+".generated", "removing stale file"))
			os.Remove(genPath)
		}
	}

	fmt.Println()
	fmt.Println(ui.Success.Render("Done!") + " Context files have been updated.")
	return nil
}

func runPreview(args []string) error {
	targetDir, format, err := parseCommonFlags(args)
	if err != nil {
		return err
	}

	absDir, err := resolveDir(targetDir)
	if err != nil {
		return err
	}

	info, err := scanAndDetect(absDir)
	if err != nil {
		return err
	}

	files := generator.Generate(info, format)
	printDryRun(files)
	return nil
}

// detectExistingFormats returns a list of formats that have existing files.
func detectExistingFormats(dir string) []generator.Format {
	var formats []generator.Format
	if analyzer.FileExists(filepath.Join(dir, "CLAUDE.md")) {
		formats = append(formats, generator.FormatClaude)
	}
	if analyzer.FileExists(filepath.Join(dir, ".cursorrules")) {
		formats = append(formats, generator.FormatCursor)
	}
	if analyzer.FileExists(filepath.Join(dir, "AGENTS.md")) {
		formats = append(formats, generator.FormatAgents)
	}
	if analyzer.FileExists(filepath.Join(dir, ".cursor", "rules", "project.mdc")) {
		formats = append(formats, generator.FormatCursorMDC)
	}
	if analyzer.FileExists(filepath.Join(dir, ".clinerules")) {
		formats = append(formats, generator.FormatCline)
	}
	if analyzer.FileExists(filepath.Join(dir, ".windsurfrules")) {
		formats = append(formats, generator.FormatWindsurf)
	}
	return formats
}

// hasFlag checks if a specific flag is present in args.
func hasFlag(args []string, names ...string) bool {
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}
	for _, a := range args {
		if nameSet[a] {
			return true
		}
	}
	return false
}

func printDryRun(files map[string]string) {
	var buf strings.Builder
	buf.WriteString("\n")
	buf.WriteString(ui.Bold.Render("Preview (dry run)"))
	buf.WriteString("\n")
	for name, content := range files {
		buf.WriteString(fmt.Sprintf("\n%s %s %s\n", ui.Warning.Render("---"), ui.Bold.Render(name), ui.Warning.Render("---")))
		buf.WriteString(content)
	}
	output := buf.String()

	lineCount := strings.Count(output, "\n")
	if ui.IsTTY() && lineCount > ui.TermHeight()-5 {
		if lessPath, err := exec.LookPath("less"); err == nil {
			cmd := exec.Command(lessPath, "-R")
			cmd.Stdin = strings.NewReader(output)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if cmd.Run() == nil {
				return
			}
		}
	}

	fmt.Print(output)
}

func printUsage() {
	fmt.Println(ui.Bold.Render("context-gen") + " - Generate AI context files for your codebase")
	fmt.Println()
	fmt.Println(ui.Bold.Render("Usage:"))
	fmt.Println("  context-gen                          Interactive mode (menu-driven)")
	fmt.Println("  context-gen init [flags]              Scan project and generate context files")
	fmt.Println("  context-gen update [flags]            Re-scan and regenerate context files")
	fmt.Println("  context-gen preview [flags]           Preview generated output")
	fmt.Println("  context-gen version                   Show version")
	fmt.Println("  context-gen help                      Show this help")
	fmt.Println()
	fmt.Println(ui.Bold.Render("Flags:"))
	fmt.Println("  -d, --dir <path>       Target directory (default: current directory)")
	fmt.Println("  -f, --format <type>    Output format: claude, cursor, agents, cursor-mdc, cline, windsurf, both, all (default: both)")
	fmt.Println("  -o, --output <path>    Output directory (default: scanned directory)")
	fmt.Println("      --dry-run          Preview without writing (init only)")
	fmt.Println("      --force            Overwrite everything, skip smart merge (update only)")
	fmt.Println()
	fmt.Println(ui.Bold.Render("Tip:") + " Just run " + ui.Bold.Render("context-gen") + " without arguments for interactive mode!")
}
