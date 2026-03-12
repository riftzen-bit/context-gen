package analyzer

// ProjectInfo holds all detected information about a project.
type ProjectInfo struct {
	RootPath     string
	Languages    []Language
	Frameworks   []string
	BuildTools   []string
	TestTools    []string
	Linters      []string
	PackageManagers map[string]string // ecosystem -> manager (e.g. "js" -> "pnpm", "python" -> "poetry")
	HasDocker    bool
	HasCI        bool
	CIProvider   string
	Structure    DirStructure
	Conventions  []Convention
	Name         string
	Description  string
	CodeStyle    CodeStyle
	Scripts      map[string]string
}

// Language represents a detected programming language with file count.
type Language struct {
	Name       string
	Extension  string
	FileCount  int
	Percentage float64
}

// DirStructure represents the project's directory layout.
type DirStructure struct {
	TopLevelDirs []string
	SubDirs      map[string][]string // parent dir -> immediate child dirs
	EntryPoints  []string            // main files, index files
	ConfigFiles  []string
	TotalFiles   int
	TotalDirs    int
}

// CodeStyle holds detected code formatting rules.
type CodeStyle struct {
	IndentStyle string   // "tabs" or "spaces"
	IndentSize  int      // 2, 4, etc.
	LineLength  int      // 80, 100, 120, etc.
	Formatter   string   // "gofmt", "prettier", "black", "rustfmt", etc.
	ExtraRules  []string
}

// Convention represents a detected coding convention.
type Convention struct {
	Category    string // "naming", "imports", "error_handling", "structure"
	Description string
	Confidence  float64 // 0.0 - 1.0
	Examples    []string
}
