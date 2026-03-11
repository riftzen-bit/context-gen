package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paul/context-gen/analyzer"
	"github.com/paul/context-gen/generator"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

func bold(s string) string   { return colorBold + s + colorReset }
func green(s string) string  { return colorGreen + s + colorReset }
func yellow(s string) string { return colorYellow + s + colorReset }
func red(s string) string    { return colorRed + s + colorReset }

func errorf(format string, a ...any) {
	fmt.Fprintf(os.Stderr, red("Error: ")+format+"\n", a...)
}

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
		fmt.Println("context-gen v0.1.0")
	default:
		errorf("Unknown command: %s", args[0])
		fmt.Println()
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		errorf("%v", err)
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
	fmt.Printf("%s %s\n", bold("Scanning"), absDir)

	scan, err := analyzer.Scan(absDir)
	if err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}
	fmt.Printf("  Found %d files in %d directories\n", scan.TotalFiles, scan.TotalDirs)

	info, err := analyzer.Detect(absDir, scan)
	if err != nil {
		return nil, fmt.Errorf("detection error: %w", err)
	}
	printDetectionSummary(info)

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
	fmt.Println(bold("Writing files"))
	for name, content := range files {
		outPath := filepath.Join(outDir, name)

		if _, err := os.Stat(outPath); err == nil {
			fmt.Printf("  %s %s already exists, writing to %s.generated\n", yellow("SKIP"), name, name)
			outPath = outPath + ".generated"
		}

		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			errorf("writing %s: %v", name, err)
			continue
		}
		fmt.Printf("  %s %s\n", green("CREATE"), outPath)
	}

	fmt.Println()
	fmt.Println(green("Done!") + " Review the generated files and customize them for your project.")
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
		fmt.Println(yellow("No existing context files found. Generating with format: " + string(format)))
	} else {
		if hasFlag(args, "--format", "-f") {
			format = formatFlag
		}
		fmt.Printf("  Detected existing format: %s\n", bold(string(format)))
	}

	info, err := scanAndDetect(absDir)
	if err != nil {
		return err
	}

	files := generator.Generate(info, format)

	// Write files (update overwrites existing)
	fmt.Println()
	fmt.Println(bold("Updating files"))
	for name, content := range files {
		outPath := filepath.Join(outDir, name)

		action := green("CREATE")
		if _, err := os.Stat(outPath); err == nil {
			action = green("UPDATE")
		}

		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			errorf("writing %s: %v", name, err)
			continue
		}
		fmt.Printf("  %s %s\n", action, outPath)
	}

	// Clean up .generated files if the real file is now written
	for name := range files {
		genPath := filepath.Join(outDir, name+".generated")
		if _, err := os.Stat(genPath); err == nil {
			fmt.Printf("  %s removing stale %s\n", yellow("CLEAN"), name+".generated")
			os.Remove(genPath)
		}
	}

	fmt.Println()
	fmt.Println(green("Done!") + " Context files have been updated.")
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
	fmt.Println(bold("Preview (dry run)"))
	for name, content := range files {
		fmt.Printf("\n%s %s %s\n", yellow("---"), bold(name), yellow("---"))
		fmt.Print(content)
	}
}

func printDetectionSummary(info *analyzer.ProjectInfo) {
	fmt.Println()
	fmt.Println(bold("Detection Results"))

	if len(info.Languages) > 0 {
		fmt.Printf("  Languages: ")
		for i, l := range info.Languages {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("%s (%.0f%%)", l.Name, l.Percentage)
		}
		fmt.Println()
	}

	if len(info.Frameworks) > 0 {
		fmt.Printf("  Frameworks: %s\n", joinSlice(info.Frameworks))
	}
	if len(info.BuildTools) > 0 {
		fmt.Printf("  Build: %s\n", joinSlice(info.BuildTools))
	}
	if len(info.PackageManagers) > 0 {
		pms := ""
		for eco, pm := range info.PackageManagers {
			if pms != "" {
				pms += ", "
			}
			pms += eco + ":" + pm
		}
		fmt.Printf("  Package Managers: %s\n", pms)
	}
	if len(info.TestTools) > 0 {
		fmt.Printf("  Testing: %s\n", joinSlice(info.TestTools))
	}
	if len(info.Linters) > 0 {
		fmt.Printf("  Linting: %s\n", joinSlice(info.Linters))
	}
}

func joinSlice(s []string) string {
	result := ""
	for i, v := range s {
		if i > 0 {
			result += ", "
		}
		result += v
	}
	return result
}

func printUsage() {
	fmt.Println(bold("context-gen") + " - Generate AI context files for your codebase")
	fmt.Println()
	fmt.Println(bold("Usage:"))
	fmt.Println("  context-gen                          Interactive mode (menu-driven)")
	fmt.Println("  context-gen init [flags]              Scan project and generate context files")
	fmt.Println("  context-gen update [flags]            Re-scan and regenerate context files")
	fmt.Println("  context-gen preview [flags]           Preview generated output")
	fmt.Println("  context-gen version                   Show version")
	fmt.Println("  context-gen help                      Show this help")
	fmt.Println()
	fmt.Println(bold("Flags:"))
	fmt.Println("  -d, --dir <path>       Target directory (default: current directory)")
	fmt.Println("  -f, --format <type>    Output format: claude, cursor, both (default: both)")
	fmt.Println("  -o, --output <path>    Output directory (default: scanned directory)")
	fmt.Println("      --dry-run          Preview without writing (init only)")
	fmt.Println()
	fmt.Println(bold("Tip:") + " Just run " + bold("context-gen") + " without arguments for interactive mode!")
}
