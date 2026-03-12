package analyzer

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// sourceExtensions are file extensions we analyze for conventions.
var sourceExtensions = map[string]bool{
	"go": true, "js": true, "ts": true, "tsx": true, "jsx": true,
	"py": true, "rs": true, "rb": true, "java": true, "kt": true,
	"swift": true, "cs": true, "cpp": true, "c": true, "php": true,
	"ex": true, "exs": true, "vue": true, "svelte": true,
}

const (
	maxFilesToSample = 20
	maxFileSize      = 50 * 1024 // 50KB
)

// DetectConventions analyzes source files to detect coding conventions.
func DetectConventions(root string, scan *ScanResult) []Convention {
	files := selectSourceFiles(scan)
	contents := readFiles(root, files)

	if len(contents) == 0 {
		return nil
	}

	var conventions []Convention

	conventions = append(conventions, detectFileNaming(files)...)
	conventions = append(conventions, detectFunctionNaming(contents)...)
	conventions = append(conventions, detectImportStyle(contents)...)
	conventions = append(conventions, detectErrorHandling(contents)...)
	conventions = append(conventions, detectTestStructure(scan)...)

	return conventions
}

// fileContent pairs a FileInfo with its source content.
type fileContent struct {
	info    FileInfo
	content string
}

// selectSourceFiles picks up to maxFilesToSample source files, preferring larger ones.
func selectSourceFiles(scan *ScanResult) []FileInfo {
	var candidates []FileInfo
	for _, f := range scan.Files {
		if !sourceExtensions[f.Extension] {
			continue
		}
		if f.Size > maxFileSize || f.Size == 0 {
			continue
		}
		candidates = append(candidates, f)
	}

	// Sort by size descending to get the most representative files.
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Size > candidates[j].Size
	})

	if len(candidates) > maxFilesToSample {
		candidates = candidates[:maxFilesToSample]
	}

	return candidates
}

// readFiles reads the content of selected files.
func readFiles(root string, files []FileInfo) []fileContent {
	var results []fileContent
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(root, f.Path))
		if err != nil {
			continue
		}
		results = append(results, fileContent{info: f, content: string(data)})
	}
	return results
}

// detectFileNaming checks whether source file names follow kebab-case, snake_case, camelCase, or PascalCase.
func detectFileNaming(files []FileInfo) []Convention {
	counts := map[string]int{
		"kebab-case": 0,
		"snake_case": 0,
		"camelCase":  0,
		"PascalCase": 0,
	}
	examples := map[string][]string{}

	for _, f := range files {
		name := strings.TrimSuffix(f.Name, "."+f.Extension)
		// Skip single-word names and test files — they don't reveal naming style.
		if !strings.ContainsAny(name, "-_") && name == strings.ToLower(name) && len(name) > 0 {
			// Single lowercase word, not distinctive enough.
			continue
		}

		style := classifyName(name)
		if style == "" {
			continue
		}
		counts[style]++
		if len(examples[style]) < 2 {
			examples[style] = append(examples[style], f.Name)
		}
	}

	total := 0
	for _, c := range counts {
		total += c
	}
	if total == 0 {
		return nil
	}

	// Find the dominant style.
	bestStyle := ""
	bestCount := 0
	for style, count := range counts {
		if count > bestCount {
			bestCount = count
			bestStyle = style
		}
	}

	confidence := float64(bestCount) / float64(total)
	if confidence < 0.5 {
		return nil
	}

	return []Convention{{
		Category:    "naming",
		Description: "Files use " + bestStyle + " naming",
		Confidence:  confidence,
		Examples:    examples[bestStyle],
	}}
}

// classifyName determines the naming convention of a string.
func classifyName(name string) string {
	if strings.Contains(name, "-") && !strings.Contains(name, "_") {
		return "kebab-case"
	}
	if strings.Contains(name, "_") && !strings.Contains(name, "-") {
		return "snake_case"
	}
	// No separator — check casing.
	if len(name) == 0 {
		return ""
	}
	if name[0] >= 'A' && name[0] <= 'Z' && strings.ContainsAny(name[1:], "abcdefghijklmnopqrstuvwxyz") {
		return "PascalCase"
	}
	if name[0] >= 'a' && name[0] <= 'z' && strings.ContainsAny(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return "camelCase"
	}
	return ""
}

var (
	// Go exported functions: func FooBar(
	goExportedFunc = regexp.MustCompile(`(?m)^func\s+([A-Z][a-zA-Z0-9]*)\s*\(`)
	// Go unexported functions: func fooBar(
	goUnexportedFunc = regexp.MustCompile(`(?m)^func\s+([a-z][a-zA-Z0-9]*)\s*\(`)
	// Go method receivers: func (r *Foo) Bar(
	goMethodExported   = regexp.MustCompile(`(?m)^func\s+\([^)]+\)\s+([A-Z][a-zA-Z0-9]*)\s*\(`)
	goMethodUnexported = regexp.MustCompile(`(?m)^func\s+\([^)]+\)\s+([a-z][a-zA-Z0-9]*)\s*\(`)

	// JS/TS: function fooBar( or const fooBar = (
	jsFuncDecl  = regexp.MustCompile(`(?m)(?:function|const|let|var)\s+([a-z][a-zA-Z0-9]*)\s*[=(]`)
	jsClassDecl = regexp.MustCompile(`(?m)class\s+([A-Z][a-zA-Z0-9]*)`)

	// Python: def foo_bar(
	pyFuncSnake  = regexp.MustCompile(`(?m)^def\s+([a-z][a-z0-9_]*)\s*\(`)
	pyFuncCamel  = regexp.MustCompile(`(?m)^def\s+([a-z][a-zA-Z0-9]*[A-Z][a-zA-Z0-9]*)\s*\(`)
	pyClassDecl  = regexp.MustCompile(`(?m)^class\s+([A-Z][a-zA-Z0-9]*)`)
)

// detectFunctionNaming detects function/method naming patterns.
func detectFunctionNaming(contents []fileContent) []Convention {
	var conventions []Convention

	goFiles := filterByExt(contents, "go")
	if len(goFiles) > 0 {
		exported := 0
		unexported := 0
		var exportedExamples, unexportedExamples []string
		for _, fc := range goFiles {
			eMatches := goExportedFunc.FindAllStringSubmatch(fc.content, -1)
			eMatches = append(eMatches, goMethodExported.FindAllStringSubmatch(fc.content, -1)...)
			for _, m := range eMatches {
				exported++
				if len(exportedExamples) < 2 {
					exportedExamples = append(exportedExamples, m[1])
				}
			}
			uMatches := goUnexportedFunc.FindAllStringSubmatch(fc.content, -1)
			uMatches = append(uMatches, goMethodUnexported.FindAllStringSubmatch(fc.content, -1)...)
			for _, m := range uMatches {
				unexported++
				if len(unexportedExamples) < 2 {
					unexportedExamples = append(unexportedExamples, m[1])
				}
			}
		}
		total := exported + unexported
		if total > 0 {
			conventions = append(conventions, Convention{
				Category:    "naming",
				Description: "Go functions follow standard naming: PascalCase for exported, camelCase for unexported",
				Confidence:  1.0,
				Examples:    dedup(append(exportedExamples, unexportedExamples...)),
			})
		}
	}

	jsFiles := filterByExt(contents, "js", "ts", "tsx", "jsx")
	if len(jsFiles) > 0 {
		camelCount := 0
		var examples []string
		for _, fc := range jsFiles {
			matches := jsFuncDecl.FindAllStringSubmatch(fc.content, -1)
			for _, m := range matches {
				camelCount++
				if len(examples) < 2 {
					examples = append(examples, m[1])
				}
			}
		}
		if camelCount > 0 {
			conventions = append(conventions, Convention{
				Category:    "naming",
				Description: "Functions use camelCase naming",
				Confidence:  1.0,
				Examples:    examples,
			})
		}
	}

	pyFiles := filterByExt(contents, "py")
	if len(pyFiles) > 0 {
		snakeCount := 0
		camelCount := 0
		var snakeExamples []string
		for _, fc := range pyFiles {
			sMatches := pyFuncSnake.FindAllStringSubmatch(fc.content, -1)
			for _, m := range sMatches {
				// Exclude camelCase matches from snake_case count.
				if !strings.ContainsAny(m[1], "ABCDEFGHIJKLMNOPQRSTUVWXYZ") && strings.Contains(m[1], "_") {
					snakeCount++
					if len(snakeExamples) < 2 {
						snakeExamples = append(snakeExamples, m[1])
					}
				}
			}
			cMatches := pyFuncCamel.FindAllStringSubmatch(fc.content, -1)
			camelCount += len(cMatches)
		}
		total := snakeCount + camelCount
		if total > 0 && snakeCount > camelCount {
			conventions = append(conventions, Convention{
				Category:    "naming",
				Description: "Python functions use snake_case naming",
				Confidence:  float64(snakeCount) / float64(total),
				Examples:    snakeExamples,
			})
		}
	}

	return conventions
}

var (
	// Go grouped imports: import (\n
	goGroupedImport   = regexp.MustCompile(`(?m)^import\s*\(`)
	goSingleImport    = regexp.MustCompile(`(?m)^import\s+"`)
	goImportBlankLine = regexp.MustCompile(`(?m)^import\s*\([^)]*\n\s*\n[^)]*\)`)

	// JS/TS relative imports: from './...' or from '../...'
	jsRelativeImport = regexp.MustCompile(`(?m)(?:import|from)\s+['"]\.\.?/[^'"]*['"]`)
	jsAbsoluteImport = regexp.MustCompile(`(?m)(?:import|from)\s+['"][^.'"][^'"]*['"]`)

	// Python relative imports: from . import or from .. import
	pyRelativeImport = regexp.MustCompile(`(?m)^from\s+\.\.?\s+import`)
	pyAbsoluteImport = regexp.MustCompile(`(?m)^(?:import|from)\s+[a-zA-Z]`)
)

// detectImportStyle detects import organization patterns.
func detectImportStyle(contents []fileContent) []Convention {
	var conventions []Convention

	// Go import grouping.
	goFiles := filterByExt(contents, "go")
	if len(goFiles) > 0 {
		grouped := 0
		single := 0
		separated := 0 // blank-line separated groups within import block
		for _, fc := range goFiles {
			if goGroupedImport.MatchString(fc.content) {
				grouped++
			}
			singles := goSingleImport.FindAllString(fc.content, -1)
			single += len(singles)

			if goImportBlankLine.MatchString(fc.content) {
				separated++
			}
		}
		total := grouped + single
		if grouped > 0 && total > 0 {
			conf := float64(grouped) / float64(total)
			c := Convention{
				Category:    "imports",
				Description: "Go imports use grouped import blocks",
				Confidence:  conf,
			}
			if separated > 0 {
				c.Description = "Go imports use grouped blocks with stdlib separated from third-party"
				c.Examples = []string{"import (\\n\\t\"fmt\"\\n\\n\\t\"github.com/...\"\\n)"}
			}
			conventions = append(conventions, c)
		}
	}

	// JS/TS import style.
	jsFiles := filterByExt(contents, "js", "ts", "tsx", "jsx")
	if len(jsFiles) > 0 {
		relCount := 0
		absCount := 0
		var relExamples, absExamples []string
		for _, fc := range jsFiles {
			rels := jsRelativeImport.FindAllString(fc.content, -1)
			relCount += len(rels)
			for _, r := range rels {
				if len(relExamples) < 1 {
					relExamples = append(relExamples, strings.TrimSpace(r))
				}
			}
			abss := jsAbsoluteImport.FindAllString(fc.content, -1)
			absCount += len(abss)
			for _, a := range abss {
				if len(absExamples) < 1 {
					absExamples = append(absExamples, strings.TrimSpace(a))
				}
			}
		}
		total := relCount + absCount
		if total > 0 {
			if relCount > absCount {
				conventions = append(conventions, Convention{
					Category:    "imports",
					Description: "Relative imports are preferred over absolute imports",
					Confidence:  float64(relCount) / float64(total),
					Examples:    relExamples,
				})
			} else if absCount > relCount {
				conventions = append(conventions, Convention{
					Category:    "imports",
					Description: "Absolute imports are preferred over relative imports",
					Confidence:  float64(absCount) / float64(total),
					Examples:    absExamples,
				})
			}
		}
	}

	// Python import style.
	pyFiles := filterByExt(contents, "py")
	if len(pyFiles) > 0 {
		relCount := 0
		absCount := 0
		for _, fc := range pyFiles {
			relCount += len(pyRelativeImport.FindAllString(fc.content, -1))
			absCount += len(pyAbsoluteImport.FindAllString(fc.content, -1))
		}
		total := relCount + absCount
		if total > 0 {
			if absCount > relCount {
				conventions = append(conventions, Convention{
					Category:    "imports",
					Description: "Python uses absolute imports",
					Confidence:  float64(absCount) / float64(total),
				})
			} else {
				conventions = append(conventions, Convention{
					Category:    "imports",
					Description: "Python uses relative imports",
					Confidence:  float64(relCount) / float64(total),
				})
			}
		}
	}

	return conventions
}

var (
	goErrCheck    = regexp.MustCompile(`(?m)if\s+err\s*!=\s*nil`)
	goErrWrap     = regexp.MustCompile(`fmt\.Errorf\([^)]*%w`)
	goSentinelErr = regexp.MustCompile(`(?m)var\s+\w+\s*=\s*(?:errors\.New|fmt\.Errorf)\(`)

	jsTryCatch   = regexp.MustCompile(`(?m)\btry\s*\{`)
	jsDotCatch   = regexp.MustCompile(`\.catch\s*\(`)
	jsAsyncAwait = regexp.MustCompile(`(?m)\basync\s+(?:function|\()`)

	pyTryExcept = regexp.MustCompile(`(?m)^\s*try\s*:`)
	pyExcept    = regexp.MustCompile(`(?m)^\s*except\s+`)
	pyBareExcept = regexp.MustCompile(`(?m)^\s*except\s*:`)
)

// detectErrorHandling detects error handling patterns.
func detectErrorHandling(contents []fileContent) []Convention {
	var conventions []Convention

	goFiles := filterByExt(contents, "go")
	if len(goFiles) > 0 {
		errChecks := 0
		wraps := 0
		sentinels := 0
		var wrapExamples []string
		for _, fc := range goFiles {
			errChecks += len(goErrCheck.FindAllString(fc.content, -1))
			wMatches := goErrWrap.FindAllString(fc.content, -1)
			wraps += len(wMatches)
			for _, w := range wMatches {
				if len(wrapExamples) < 2 {
					wrapExamples = append(wrapExamples, strings.TrimSpace(w))
				}
			}
			sentinels += len(goSentinelErr.FindAllString(fc.content, -1))
		}
		if errChecks > 0 {
			conventions = append(conventions, Convention{
				Category:    "error_handling",
				Description: "Go uses standard if err != nil error checking",
				Confidence:  1.0,
				Examples:    []string{"if err != nil { return err }"},
			})
		}
		if wraps > 0 && errChecks > 0 {
			conf := float64(wraps) / float64(errChecks)
			if conf > 1.0 {
				conf = 1.0
			}
			conventions = append(conventions, Convention{
				Category:    "error_handling",
				Description: "Errors are wrapped with fmt.Errorf and %w for context",
				Confidence:  conf,
				Examples:    wrapExamples,
			})
		}
		if sentinels > 0 {
			conventions = append(conventions, Convention{
				Category:    "error_handling",
				Description: "Sentinel errors are defined with errors.New or fmt.Errorf",
				Confidence:  0.8,
			})
		}
	}

	jsFiles := filterByExt(contents, "js", "ts", "tsx", "jsx")
	if len(jsFiles) > 0 {
		tryCatches := 0
		dotCatches := 0
		asyncFns := 0
		for _, fc := range jsFiles {
			tryCatches += len(jsTryCatch.FindAllString(fc.content, -1))
			dotCatches += len(jsDotCatch.FindAllString(fc.content, -1))
			asyncFns += len(jsAsyncAwait.FindAllString(fc.content, -1))
		}
		if asyncFns > 0 && tryCatches > 0 {
			conventions = append(conventions, Convention{
				Category:    "error_handling",
				Description: "async/await with try/catch for error handling",
				Confidence:  1.0,
				Examples:    []string{"try { await ... } catch (err) { ... }"},
			})
		} else if tryCatches > dotCatches && tryCatches > 0 {
			conventions = append(conventions, Convention{
				Category:    "error_handling",
				Description: "try/catch blocks for error handling",
				Confidence:  float64(tryCatches) / float64(tryCatches+dotCatches),
				Examples:    []string{"try { ... } catch (err) { ... }"},
			})
		} else if dotCatches > 0 {
			conventions = append(conventions, Convention{
				Category:    "error_handling",
				Description: "Promise .catch() for error handling",
				Confidence:  float64(dotCatches) / float64(tryCatches+dotCatches),
				Examples:    []string{"promise.catch(err => { ... })"},
			})
		}
	}

	pyFiles := filterByExt(contents, "py")
	if len(pyFiles) > 0 {
		tries := 0
		specificExcepts := 0
		bareExcepts := 0
		for _, fc := range pyFiles {
			tries += len(pyTryExcept.FindAllString(fc.content, -1))
			specificExcepts += len(pyExcept.FindAllString(fc.content, -1))
			bareExcepts += len(pyBareExcept.FindAllString(fc.content, -1))
		}
		// Specific excepts count includes bare excepts, so subtract.
		specificExcepts -= bareExcepts
		if tries > 0 {
			if specificExcepts > bareExcepts {
				conventions = append(conventions, Convention{
					Category:    "error_handling",
					Description: "Python uses try/except with specific exception types",
					Confidence:  float64(specificExcepts) / float64(specificExcepts+bareExcepts),
					Examples:    []string{"except ValueError as e:"},
				})
			} else {
				conventions = append(conventions, Convention{
					Category:    "error_handling",
					Description: "Python uses try/except for error handling",
					Confidence:  0.7,
				})
			}
		}
	}

	return conventions
}

// detectTestStructure detects whether tests live alongside source or in separate directories.
func detectTestStructure(scan *ScanResult) []Convention {
	testsAlongside := 0
	testsInTestDir := 0
	var alongsideExamples, testDirExamples []string

	for _, f := range scan.Files {
		isTest := false
		switch {
		case strings.HasSuffix(f.Name, "_test.go"):
			isTest = true
		case strings.HasSuffix(f.Name, ".test.js"),
			strings.HasSuffix(f.Name, ".test.ts"),
			strings.HasSuffix(f.Name, ".test.tsx"),
			strings.HasSuffix(f.Name, ".test.jsx"),
			strings.HasSuffix(f.Name, ".spec.js"),
			strings.HasSuffix(f.Name, ".spec.ts"),
			strings.HasSuffix(f.Name, ".spec.tsx"),
			strings.HasSuffix(f.Name, ".spec.jsx"):
			isTest = true
		case strings.HasPrefix(f.Name, "test_") && f.Extension == "py":
			isTest = true
		}

		if !isTest {
			continue
		}

		dir := filepath.Dir(f.Path)
		parts := strings.Split(dir, string(filepath.Separator))
		inTestDir := false
		for _, p := range parts {
			if p == "test" || p == "tests" || p == "__tests__" || p == "spec" {
				inTestDir = true
				break
			}
		}

		if inTestDir {
			testsInTestDir++
			if len(testDirExamples) < 2 {
				testDirExamples = append(testDirExamples, f.Path)
			}
		} else {
			testsAlongside++
			if len(alongsideExamples) < 2 {
				alongsideExamples = append(alongsideExamples, f.Path)
			}
		}
	}

	total := testsAlongside + testsInTestDir
	if total == 0 {
		return nil
	}

	if testsAlongside >= testsInTestDir {
		return []Convention{{
			Category:    "structure",
			Description: "Tests are co-located alongside source files",
			Confidence:  float64(testsAlongside) / float64(total),
			Examples:    alongsideExamples,
		}}
	}

	return []Convention{{
		Category:    "structure",
		Description: "Tests are in separate test directories",
		Confidence:  float64(testsInTestDir) / float64(total),
		Examples:    testDirExamples,
	}}
}

// filterByExt returns file contents matching any of the given extensions.
func filterByExt(contents []fileContent, exts ...string) []fileContent {
	extSet := make(map[string]bool, len(exts))
	for _, e := range exts {
		extSet[e] = true
	}
	var result []fileContent
	for _, fc := range contents {
		if extSet[fc.info.Extension] {
			result = append(result, fc)
		}
	}
	return result
}

// dedup returns a slice with duplicate strings removed.
func dedup(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
