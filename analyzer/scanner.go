package analyzer

import (
	"os"
	"path/filepath"
	"strings"
)

// Default directories to skip during scanning.
var skipDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".next":        true,
	"dist":         true,
	"build":        true,
	"target":       true,
	"__pycache__":  true,
	".venv":        true,
	"venv":         true,
	".idea":        true,
	".vscode":      true,
	"coverage":     true,
	".cache":       true,
	"bin":          true,
	"obj":          true,
}

// ScanResult holds raw scan data before analysis.
type ScanResult struct {
	Files      []FileInfo
	Dirs       []string
	TotalFiles int
	TotalDirs  int
}

// FileInfo holds metadata about a single file.
type FileInfo struct {
	Path      string
	Name      string
	Extension string
	Size      int64
	IsConfig  bool
}

// configFiles are filenames that indicate project configuration.
var configFiles = map[string]bool{
	"package.json":      true,
	"go.mod":            true,
	"Cargo.toml":        true,
	"pyproject.toml":    true,
	"setup.py":          true,
	"requirements.txt":  true,
	"pom.xml":           true,
	"build.gradle":      true,
	"Makefile":          true,
	"Dockerfile":        true,
	"docker-compose.yml": true,
	"docker-compose.yaml": true,
	".eslintrc.json":    true,
	".eslintrc.js":      true,
	"eslint.config.js":  true,
	"eslint.config.mjs": true,
	"tsconfig.json":     true,
	"vite.config.ts":    true,
	"vite.config.js":    true,
	"webpack.config.js": true,
	"next.config.js":    true,
	"next.config.mjs":   true,
	"next.config.ts":    true,
	"tailwind.config.js": true,
	"tailwind.config.ts": true,
	"jest.config.js":    true,
	"jest.config.ts":    true,
	"vitest.config.ts":  true,
	".prettierrc":       true,
	".prettierrc.json":  true,
	"biome.json":        true,
	"deno.json":         true,
	"bun.lockb":         true,
	"Gemfile":           true,
	"composer.json":     true,
	"mix.exs":           true,
	"pubspec.yaml":      true,
	"CMakeLists.txt":    true,
}

// Scan walks the project directory and collects file metadata.
func Scan(root string) (*ScanResult, error) {
	result := &ScanResult{}

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip files we can't read
		}

		rel, _ := filepath.Rel(root, path)

		if d.IsDir() {
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			if rel != "." {
				result.Dirs = append(result.Dirs, rel)
				result.TotalDirs++
			}
			return nil
		}

		// Skip hidden files (except config dotfiles)
		if strings.HasPrefix(d.Name(), ".") && !configFiles[d.Name()] {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		ext := strings.TrimPrefix(filepath.Ext(d.Name()), ".")

		result.Files = append(result.Files, FileInfo{
			Path:      rel,
			Name:      d.Name(),
			Extension: ext,
			Size:      info.Size(),
			IsConfig:  configFiles[d.Name()],
		})
		result.TotalFiles++

		return nil
	})

	return result, err
}
