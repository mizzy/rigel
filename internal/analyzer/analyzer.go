package analyzer

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mizzy/rigel/internal/llm"
)

type RepoAnalyzer struct {
	rootPath string
	files    []FileInfo
	dirs     []string
	provider llm.Provider
}

type FileInfo struct {
	Path         string
	RelativePath string
	Extension    string
	Size         int64
	IsTest       bool
}

func NewRepoAnalyzer(provider llm.Provider) *RepoAnalyzer {
	cwd, _ := os.Getwd()
	return &RepoAnalyzer{
		rootPath: cwd,
		files:    []FileInfo{},
		dirs:     []string{},
		provider: provider,
	}
}

func (r *RepoAnalyzer) Analyze() (string, error) {
	// Walk through the repository
	err := filepath.Walk(r.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files that can't be accessed
		}

		relPath, _ := filepath.Rel(r.rootPath, path)

		// Skip hidden directories and vendor/node_modules
		if info.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") && base != "." {
				return filepath.SkipDir
			}
			if base == "vendor" || base == "node_modules" || base == ".git" {
				return filepath.SkipDir
			}
			if path != r.rootPath {
				r.dirs = append(r.dirs, relPath)
			}
			return nil
		}

		// Skip hidden files and binary files
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		ext := filepath.Ext(path)
		// Only include source code files
		if isSourceFile(ext) {
			r.files = append(r.files, FileInfo{
				Path:         path,
				RelativePath: relPath,
				Extension:    ext,
				Size:         info.Size(),
				IsTest:       strings.Contains(path, "_test.") || strings.Contains(path, ".test.") || strings.Contains(path, ".spec."),
			})
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	// Generate the AGENTS.md content using LLM
	return r.generateAgentsContentWithLLM()
}

func (r *RepoAnalyzer) generateAgentsContentWithLLM() (string, error) {
	// Collect repository information
	info := r.collectRepositoryInfo()

	// Create prompt for LLM
	prompt := fmt.Sprintf(`Analyze the following repository structure and generate an AGENTS.md file that provides an AI-friendly overview of the codebase.

Repository Information:
%s

Please create a comprehensive AGENTS.md file that includes:
1. Repository Overview - project name, type, and purpose
2. Main Components - table with package names and their purposes
3. Key Files - table with important files and why they matter
4. Development information - how to test, build, and contribute
5. Architecture diagram if appropriate
6. Any other relevant information for AI agents to understand the codebase

Format the output as a proper Markdown file starting with "# AGENTS.md".
Make sure the content is well-structured, informative, and helps AI agents understand the codebase quickly.`, info)

	// Generate content using LLM
	ctx := context.Background()
	response, err := r.provider.Generate(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate AGENTS.md content: %w", err)
	}

	return response, nil
}

func (r *RepoAnalyzer) collectRepositoryInfo() string {
	var sb strings.Builder

	// Add directory structure (simplified)
	sb.WriteString("Directory Structure:\n")
	sb.WriteString("```\n")
	r.writeSimplifiedTree(&sb)
	sb.WriteString("```\n\n")

	// Add key Go files only (limit to important ones)
	sb.WriteString("Key Files:\n")
	count := 0
	for _, file := range r.files {
		if file.Extension == ".go" && !file.IsTest && count < 20 {
			sb.WriteString(fmt.Sprintf("- %s\n", file.RelativePath))
			count++
		}
	}
	if count == 20 {
		sb.WriteString("- ... (more files)\n")
	}
	sb.WriteString("\n")

	// Add statistics
	stats := r.calculateStats()
	sb.WriteString("Statistics:\n")
	sb.WriteString(fmt.Sprintf("- Total Go files: %d\n", stats.GoFiles))
	sb.WriteString(fmt.Sprintf("- Test files: %d\n", stats.TestFiles))
	sb.WriteString(fmt.Sprintf("- Estimated lines of code: %d\n", stats.EstimatedLOC))

	return sb.String()
}

func (r *RepoAnalyzer) writeSimplifiedTree(sb *strings.Builder) {
	// Create a simple tree showing main directories only
	sb.WriteString("rigel/\n")
	mainDirs := make(map[string]bool)
	for _, dir := range r.dirs {
		parts := strings.Split(dir, string(os.PathSeparator))
		if len(parts) > 0 {
			mainDirs[parts[0]] = true
		}
	}

	sortedDirs := make([]string, 0, len(mainDirs))
	for dir := range mainDirs {
		sortedDirs = append(sortedDirs, dir)
	}
	sort.Strings(sortedDirs)

	for i, dir := range sortedDirs {
		if i == len(sortedDirs)-1 {
			sb.WriteString(fmt.Sprintf("└── %s/\n", dir))
		} else {
			sb.WriteString(fmt.Sprintf("├── %s/\n", dir))
		}
	}
}

type Stats struct {
	GoFiles      int
	TestFiles    int
	EstimatedLOC int
}

func (r *RepoAnalyzer) calculateStats() Stats {
	stats := Stats{}
	for _, file := range r.files {
		if file.Extension == ".go" {
			stats.GoFiles++
			if file.IsTest {
				stats.TestFiles++
			}
			// Rough estimate: 40 lines per KB
			stats.EstimatedLOC += int(file.Size / 25)
		}
	}
	return stats
}

func (r *RepoAnalyzer) WriteAgentsFile(content string) error {
	filePath := filepath.Join(r.rootPath, "AGENTS.md")
	return os.WriteFile(filePath, []byte(content), 0644)
}

func isSourceFile(ext string) bool {
	sourceExts := []string{
		".go", ".mod", ".sum",
		".js", ".jsx", ".ts", ".tsx",
		".py", ".rb", ".java", ".c", ".cpp", ".h",
		".rs", ".swift", ".kt", ".scala",
		".sh", ".bash", ".zsh",
		".yml", ".yaml", ".json", ".toml", ".xml",
		".md", ".txt", ".rst",
		".sql", ".proto",
		".Dockerfile", ".dockerignore",
		".gitignore", ".env.example",
	}

	for _, sExt := range sourceExts {
		if ext == sExt {
			return true
		}
	}

	// Check for files without extensions (like Dockerfile, Makefile)
	if ext == "" {
		return true
	}

	return false
}
