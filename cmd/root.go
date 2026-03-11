package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paul/context-gen/analyzer"
	"github.com/paul/context-gen/generator"
	"github.com/paul/context-gen/ui"
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
				default:
					return "", "", fmt.Errorf("unknown format: %s (use: claude, cursor, both)", args[i+1])
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

	absDir, err := resolveDir(targetDir)
	if err != nil {
		return err
	}

	outDir, err := resolveOutputDir(outputDir, absDir)
	if err != nil {
		return err
	}

	// Auto-detect format from existing files if no explicit --format given
	format := detectExistingFormat(outDir)
	if format == "" {
		format = formatFlag
		fmt.Println(ui.Warning.Render("No existing context files found. Generating with format: " + string(format)))
	} else {
		if hasFlag(args, "--format", "-f") {
			format = formatFlag
		}
		fmt.Printf("  Detected existing format: %s\n", ui.Bold.Render(string(format)))
	}

	info, err := scanAndDetect(absDir)
	if err != nil {
		return err
	}

	files := generator.Generate(info, format)

	// Write files (update overwrites existing)
	fmt.Println()
	fmt.Println(ui.Bold.Render("Updating files"))
	for name, content := range files {
		outPath := filepath.Join(outDir, name)

		action := ui.FormatFileCreated
		if _, err := os.Stat(outPath); err == nil {
			action = ui.FormatFileUpdated
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

// detectExistingFormat checks which context files exist in the directory.
func detectExistingFormat(dir string) generator.Format {
	hasClaude := fileExists(filepath.Join(dir, "CLAUDE.md"))
	hasCursor := fileExists(filepath.Join(dir, ".cursorrules"))

	switch {
	case hasClaude && hasCursor:
		return generator.FormatBoth
	case hasClaude:
		return generator.FormatClaude
	case hasCursor:
		return generator.FormatCursor
	default:
		return "" // no existing files
	}
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func printDryRun(files map[string]string) {
	fmt.Println()
	fmt.Println(ui.Bold.Render("Preview (dry run)"))
	for name, content := range files {
		fmt.Printf("\n%s %s %s\n", ui.Warning.Render("---"), ui.Bold.Render(name), ui.Warning.Render("---"))
		fmt.Print(content)
	}
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
	fmt.Println("  -f, --format <type>    Output format: claude, cursor, both (default: both)")
	fmt.Println("  -o, --output <path>    Output directory (default: scanned directory)")
	fmt.Println("      --dry-run          Preview without writing (init only)")
	fmt.Println()
	fmt.Println(ui.Bold.Render("Tip:") + " Just run " + ui.Bold.Render("context-gen") + " without arguments for interactive mode!")
}
