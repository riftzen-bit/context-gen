package analyzer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// languageMap maps file extensions to language names.
var languageMap = map[string]string{
	"go":    "Go",
	"rs":    "Rust",
	"py":    "Python",
	"js":    "JavaScript",
	"ts":    "TypeScript",
	"tsx":   "TypeScript (React)",
	"jsx":   "JavaScript (React)",
	"rb":    "Ruby",
	"java":  "Java",
	"kt":    "Kotlin",
	"swift": "Swift",
	"cs":    "C#",
	"cpp":   "C++",
	"c":     "C",
	"php":   "PHP",
	"ex":    "Elixir",
	"exs":   "Elixir",
	"dart":  "Dart",
	"vue":   "Vue",
	"svelte": "Svelte",
	"zig":   "Zig",
	"lua":   "Lua",
	"sh":    "Shell",
	"bash":  "Shell",
	"sql":   "SQL",
}

// Detect analyzes scan results and builds a ProjectInfo.
func Detect(root string, scan *ScanResult) (*ProjectInfo, error) {
	info := &ProjectInfo{
		RootPath:        root,
		PackageManagers: make(map[string]string),
	}

	info.Languages = detectLanguages(scan)
	info.Structure = buildStructure(scan)

	detectFromConfigs(root, scan, info)
	detectCI(scan, info)
	detectDocker(scan, info)

	info.Conventions = DetectConventions(root, scan)

	return info, nil
}

func detectLanguages(scan *ScanResult) []Language {
	counts := make(map[string]int)
	extMap := make(map[string]string) // ext -> language name

	for _, f := range scan.Files {
		if f.IsConfig || f.Extension == "" {
			continue
		}
		if name, ok := languageMap[f.Extension]; ok {
			counts[name]++
			extMap[name] = f.Extension
		}
	}

	total := 0
	for _, c := range counts {
		total += c
	}

	var langs []Language
	for name, count := range counts {
		langs = append(langs, Language{
			Name:       name,
			Extension:  extMap[name],
			FileCount:  count,
			Percentage: float64(count) / float64(total) * 100,
		})
	}

	sort.Slice(langs, func(i, j int) bool {
		return langs[i].FileCount > langs[j].FileCount
	})

	return langs
}

func buildStructure(scan *ScanResult) DirStructure {
	ds := DirStructure{
		TotalFiles: scan.TotalFiles,
		TotalDirs:  scan.TotalDirs,
	}

	topLevel := make(map[string]bool)
	for _, d := range scan.Dirs {
		parts := strings.SplitN(d, string(filepath.Separator), 2)
		topLevel[parts[0]] = true
	}
	for d := range topLevel {
		ds.TopLevelDirs = append(ds.TopLevelDirs, d)
	}
	sort.Strings(ds.TopLevelDirs)

	entryNames := map[string]bool{
		"main.go": true, "main.py": true, "main.rs": true,
		"index.js": true, "index.ts": true, "index.tsx": true,
		"app.js": true, "app.ts": true, "app.tsx": true,
		"App.tsx": true, "App.jsx": true,
		"manage.py": true, "server.go": true, "server.ts": true,
	}
	for _, f := range scan.Files {
		if entryNames[f.Name] {
			ds.EntryPoints = append(ds.EntryPoints, f.Path)
		}
	}

	for _, f := range scan.Files {
		if f.IsConfig {
			ds.ConfigFiles = append(ds.ConfigFiles, f.Path)
		}
	}

	return ds
}

// packageJSON is a minimal representation for detection purposes.
type packageJSON struct {
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
	Scripts         map[string]string `json:"scripts"`
}

func detectFromConfigs(root string, scan *ScanResult, info *ProjectInfo) {
	for _, f := range scan.Files {
		if !f.IsConfig {
			continue
		}

		switch f.Name {
		case "go.mod":
			info.BuildTools = appendUnique(info.BuildTools, "Go Modules")

		case "Cargo.toml":
			info.BuildTools = appendUnique(info.BuildTools, "Cargo")

		case "pyproject.toml":
			info.BuildTools = appendUnique(info.BuildTools, "pyproject")
			data, err := os.ReadFile(filepath.Join(root, f.Path))
			if err == nil {
				content := string(data)
				if strings.Contains(content, "[tool.poetry]") {
					info.PackageManagers["python"] = "poetry"
				} else if strings.Contains(content, "[tool.pdm]") {
					info.PackageManagers["python"] = "pdm"
				} else if strings.Contains(content, "[project]") {
					info.PackageManagers["python"] = "pip"
				}
				if strings.Contains(content, "[tool.pytest]") || strings.Contains(content, "pytest") {
					info.TestTools = appendUnique(info.TestTools, "pytest")
				}
				if strings.Contains(content, "ruff") {
					info.Linters = appendUnique(info.Linters, "ruff")
				}
				if strings.Contains(content, "mypy") {
					info.Linters = appendUnique(info.Linters, "mypy")
				}
			}

		case "requirements.txt", "setup.py":
			if info.PackageManagers["python"] == "" {
				info.PackageManagers["python"] = "pip"
			}

		case "package.json":
			detectFromPackageJSON(root, f.Path, info)

		case "Makefile":
			info.BuildTools = appendUnique(info.BuildTools, "Make")

		case "pom.xml":
			info.BuildTools = appendUnique(info.BuildTools, "Maven")

		case "build.gradle":
			info.BuildTools = appendUnique(info.BuildTools, "Gradle")

		case "Gemfile":
			info.PackageManagers["ruby"] = "bundler"
			info.BuildTools = appendUnique(info.BuildTools, "Bundler")

		case "composer.json":
			info.PackageManagers["php"] = "composer"

		case "mix.exs":
			info.BuildTools = appendUnique(info.BuildTools, "Mix")

		case "pubspec.yaml":
			info.PackageManagers["dart"] = "pub"
			info.BuildTools = appendUnique(info.BuildTools, "pub")

		case "CMakeLists.txt":
			info.BuildTools = appendUnique(info.BuildTools, "CMake")

		case "biome.json":
			info.Linters = appendUnique(info.Linters, "Biome")

		case "deno.json":
			info.BuildTools = appendUnique(info.BuildTools, "Deno")
			info.PackageManagers["js"] = "deno"
		}
	}
}

func detectFromPackageJSON(root, path string, info *ProjectInfo) {
	data, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		return
	}

	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return
	}

	// Detect JS package manager
	if info.PackageManagers["js"] == "" {
		if fileExists(filepath.Join(root, "bun.lockb")) {
			info.PackageManagers["js"] = "bun"
		} else if fileExists(filepath.Join(root, "pnpm-lock.yaml")) {
			info.PackageManagers["js"] = "pnpm"
		} else if fileExists(filepath.Join(root, "yarn.lock")) {
			info.PackageManagers["js"] = "yarn"
		} else {
			info.PackageManagers["js"] = "npm"
		}
	}

	allDeps := mergeMaps(pkg.Dependencies, pkg.DevDependencies)

	// Frameworks
	frameworkChecks := map[string]string{
		"next":        "Next.js",
		"react":       "React",
		"vue":         "Vue.js",
		"nuxt":        "Nuxt",
		"svelte":      "Svelte",
		"@angular/core": "Angular",
		"express":     "Express",
		"fastify":     "Fastify",
		"hono":        "Hono",
		"remix":       "Remix",
		"astro":       "Astro",
		"gatsby":      "Gatsby",
		"electron":    "Electron",
	}
	for dep, name := range frameworkChecks {
		if _, ok := allDeps[dep]; ok {
			info.Frameworks = appendUnique(info.Frameworks, name)
		}
	}

	// Build tools
	buildChecks := map[string]string{
		"vite":    "Vite",
		"webpack": "Webpack",
		"esbuild": "esbuild",
		"rollup":  "Rollup",
		"turbo":   "Turborepo",
		"tsup":    "tsup",
		"swc":     "SWC",
	}
	for dep, name := range buildChecks {
		if _, ok := allDeps[dep]; ok {
			info.BuildTools = appendUnique(info.BuildTools, name)
		}
	}

	// Test tools
	testChecks := map[string]string{
		"jest":     "Jest",
		"vitest":   "Vitest",
		"mocha":    "Mocha",
		"cypress":  "Cypress",
		"playwright": "Playwright",
		"@testing-library/react": "Testing Library",
	}
	for dep, name := range testChecks {
		if _, ok := allDeps[dep]; ok {
			info.TestTools = appendUnique(info.TestTools, name)
		}
	}

	// Linters
	lintChecks := map[string]string{
		"eslint":     "ESLint",
		"prettier":   "Prettier",
		"@biomejs/biome": "Biome",
		"stylelint":  "Stylelint",
	}
	for dep, name := range lintChecks {
		if _, ok := allDeps[dep]; ok {
			info.Linters = appendUnique(info.Linters, name)
		}
	}

	// CSS framework detection
	cssChecks := map[string]string{
		"tailwindcss":       "Tailwind CSS",
		"styled-components": "styled-components",
		"@emotion/react":    "Emotion",
		"sass":              "Sass",
	}
	for dep, name := range cssChecks {
		if _, ok := allDeps[dep]; ok {
			info.Frameworks = appendUnique(info.Frameworks, name)
		}
	}
}

func detectCI(scan *ScanResult, info *ProjectInfo) {
	for _, d := range scan.Dirs {
		switch {
		case d == ".github" || strings.HasPrefix(d, ".github"+string(filepath.Separator)):
			info.HasCI = true
			info.CIProvider = "GitHub Actions"
		case d == ".gitlab-ci.yml":
			info.HasCI = true
			info.CIProvider = "GitLab CI"
		case d == ".circleci":
			info.HasCI = true
			info.CIProvider = "CircleCI"
		}
	}

	for _, f := range scan.Files {
		switch f.Name {
		case ".travis.yml":
			info.HasCI = true
			info.CIProvider = "Travis CI"
		case "Jenkinsfile":
			info.HasCI = true
			info.CIProvider = "Jenkins"
		case ".gitlab-ci.yml":
			info.HasCI = true
			info.CIProvider = "GitLab CI"
		}
	}
}

func detectDocker(scan *ScanResult, info *ProjectInfo) {
	for _, f := range scan.Files {
		if f.Name == "Dockerfile" || f.Name == "docker-compose.yml" || f.Name == "docker-compose.yaml" {
			info.HasDocker = true
			return
		}
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

func mergeMaps(a, b map[string]string) map[string]string {
	result := make(map[string]string)
	for k, v := range a {
		result[k] = v
	}
	for k, v := range b {
		result[k] = v
	}
	return result
}
