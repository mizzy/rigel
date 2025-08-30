package tui

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mizzy/rigel/internal/llm"
)

// RepoAnalyzer analyzes repository structure and content
type RepoAnalyzer struct {
	provider llm.Provider
	rootPath string
}

// NewRepoAnalyzer creates a new repository analyzer
func NewRepoAnalyzer(provider llm.Provider, rootPath string) *RepoAnalyzer {
	if rootPath == "" {
		rootPath, _ = os.Getwd()
	}
	return &RepoAnalyzer{
		provider: provider,
		rootPath: rootPath,
	}
}

// AnalyzeRepository analyzes the repository and generates AGENTS.md
func (r *RepoAnalyzer) AnalyzeRepository(ctx context.Context) error {
	// Collect repository information
	info, err := r.collectRepoInfo()
	if err != nil {
		return fmt.Errorf("failed to collect repository info: %w", err)
	}

	// Generate summary using LLM
	summary, err := r.generateSummary(ctx, info)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	// Write AGENTS.md
	if err := r.writeAgentsMD(summary); err != nil {
		return fmt.Errorf("failed to write AGENTS.md: %w", err)
	}

	return nil
}

// RepoInfo contains repository information
type RepoInfo struct {
	Name        string
	Description string
	MainFiles   []FileInfo
	Structure   string
	Languages   map[string]int
	TotalFiles  int
	TotalLines  int
}

// FileInfo contains file information
type FileInfo struct {
	Path    string
	Size    int64
	Lines   int
	Summary string
}

func (r *RepoAnalyzer) collectRepoInfo() (*RepoInfo, error) {
	info := &RepoInfo{
		Name:      filepath.Base(r.rootPath),
		MainFiles: []FileInfo{},
		Languages: make(map[string]int),
	}

	// Read README if exists
	readmePath := filepath.Join(r.rootPath, "README.md")
	if content, err := os.ReadFile(readmePath); err == nil {
		lines := strings.Split(string(content), "\n")
		if len(lines) > 0 {
			info.Description = strings.TrimPrefix(lines[0], "# ")
		}
	}

	// Analyze directory structure
	structure, err := r.analyzeStructure()
	if err != nil {
		return nil, err
	}
	info.Structure = structure

	// Collect main files
	err = filepath.WalkDir(r.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories and vendor
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip non-code files
		ext := filepath.Ext(path)
		if !isCodeFile(ext) {
			return nil
		}

		// Count language
		lang := getLanguageFromExt(ext)
		info.Languages[lang]++

		// Get file info
		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil
		}

		// Read file for important files
		relPath, _ := filepath.Rel(r.rootPath, path)
		if isImportantFile(relPath) && fileInfo.Size() < 100000 { // Skip large files
			content, err := os.ReadFile(path)
			if err == nil {
				lines := strings.Count(string(content), "\n")
				info.TotalLines += lines

				info.MainFiles = append(info.MainFiles, FileInfo{
					Path:  relPath,
					Size:  fileInfo.Size(),
					Lines: lines,
				})
			}
		}

		info.TotalFiles++
		return nil
	})

	if err != nil {
		return nil, err
	}

	return info, nil
}

func (r *RepoAnalyzer) analyzeStructure() (string, error) {
	var dirs []string

	err := filepath.WalkDir(r.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			relPath, _ := filepath.Rel(r.rootPath, path)
			if relPath == "." {
				return nil
			}

			// Skip hidden and vendor directories
			parts := strings.Split(relPath, string(filepath.Separator))
			for _, part := range parts {
				if strings.HasPrefix(part, ".") || part == "vendor" || part == "node_modules" {
					return filepath.SkipDir
				}
			}

			// Only include top 2 levels
			if len(parts) <= 2 {
				dirs = append(dirs, relPath)
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	sort.Strings(dirs)

	// Build tree structure
	var tree strings.Builder
	tree.WriteString(".\n")
	for _, dir := range dirs {
		level := strings.Count(dir, string(filepath.Separator))
		indent := strings.Repeat("  ", level)
		name := filepath.Base(dir)
		tree.WriteString(fmt.Sprintf("%s└── %s/\n", indent, name))
	}

	return tree.String(), nil
}

func (r *RepoAnalyzer) generateSummary(ctx context.Context, info *RepoInfo) (string, error) {
	// Prepare prompt for LLM
	prompt := fmt.Sprintf(`Analyze this repository and generate a comprehensive AGENTS.md file.

Repository: %s
Description: %s

Structure:
%s

Languages: %v
Total Files: %d
Total Lines: %d

Main Files:
`, info.Name, info.Description, info.Structure, info.Languages, info.TotalFiles, info.TotalLines)

	for _, file := range info.MainFiles {
		prompt += fmt.Sprintf("- %s (%d lines)\n", file.Path, file.Lines)
	}

	prompt += `
Generate an AGENTS.md file with the following sections:
1. Repository Overview - Brief description of what this repository does
2. Main Components - List and describe the main components/modules
3. Key Files - Important files and their purposes
4. Architecture - High-level architecture description
5. Technologies Used - Main technologies and frameworks
6. Getting Started - Quick start guide for AI agents working with this codebase
7. Development Guidelines - Key patterns and conventions to follow

Format the output as proper Markdown. Be comprehensive but concise.`

	response, err := r.provider.Generate(ctx, prompt)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (r *RepoAnalyzer) writeAgentsMD(content string) error {
	agentsPath := filepath.Join(r.rootPath, "AGENTS.md")

	// Add header if not present
	if !strings.HasPrefix(content, "# ") {
		content = fmt.Sprintf("# AI Agent Guide for %s\n\n%s", filepath.Base(r.rootPath), content)
	}

	// Add footer
	content += "\n\n---\n\n*This file was automatically generated by Rigel AI to help AI agents understand this codebase.*\n"

	return os.WriteFile(agentsPath, []byte(content), 0644)
}

// Helper functions

func isCodeFile(ext string) bool {
	codeExts := []string{
		".go", ".py", ".js", ".ts", ".jsx", ".tsx",
		".java", ".c", ".cpp", ".h", ".hpp", ".cs",
		".rb", ".php", ".rs", ".swift", ".kt", ".scala",
		".sh", ".bash", ".zsh", ".fish",
		".yml", ".yaml", ".json", ".xml", ".toml",
	}

	for _, e := range codeExts {
		if ext == e {
			return true
		}
	}
	return false
}

func getLanguageFromExt(ext string) string {
	langMap := map[string]string{
		".go":    "Go",
		".py":    "Python",
		".js":    "JavaScript",
		".ts":    "TypeScript",
		".jsx":   "JavaScript",
		".tsx":   "TypeScript",
		".java":  "Java",
		".c":     "C",
		".cpp":   "C++",
		".h":     "C",
		".hpp":   "C++",
		".cs":    "C#",
		".rb":    "Ruby",
		".php":   "PHP",
		".rs":    "Rust",
		".swift": "Swift",
		".kt":    "Kotlin",
		".scala": "Scala",
	}

	if lang, ok := langMap[ext]; ok {
		return lang
	}
	return "Other"
}

func isImportantFile(path string) bool {
	importantPatterns := []string{
		"main.go", "main.py", "index.js", "index.ts",
		"app.js", "app.py", "server.go", "server.js",
		"README", "Makefile", "Dockerfile",
		"package.json", "go.mod", "requirements.txt",
		"pom.xml", "build.gradle", "Cargo.toml",
	}

	base := filepath.Base(path)
	for _, pattern := range importantPatterns {
		if strings.Contains(base, pattern) {
			return true
		}
	}

	// Check if in important directories
	importantDirs := []string{"cmd/", "src/", "internal/", "pkg/", "lib/", "app/"}
	for _, dir := range importantDirs {
		if strings.Contains(path, dir) {
			return true
		}
	}

	return false
}
