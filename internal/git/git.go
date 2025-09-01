package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Info holds git repository information
type Info struct {
	RepoName string
	Branch   string
}

// GetRepoInfo returns the current git repository information
func GetRepoInfo() *Info {
	info := &Info{}

	// Get repository name
	if repoName := getRepoName(); repoName != "" {
		info.RepoName = repoName
	}

	// Get current branch
	if branch := getCurrentBranch(); branch != "" {
		info.Branch = branch
	}

	// Return nil if we couldn't get any git info
	if info.RepoName == "" && info.Branch == "" {
		return nil
	}

	return info
}

// getRepoName extracts the repository name from git remote or directory
func getRepoName() string {
	// Try to get from git remote
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err == nil {
		url := strings.TrimSpace(string(output))
		if url != "" {
			// Extract owner/repo from URL
			// Handle both SSH and HTTPS URLs
			url = strings.TrimSuffix(url, ".git")

			// Handle SSH format (git@github.com:owner/repo)
			if strings.Contains(url, "@") && strings.Contains(url, ":") {
				parts := strings.Split(url, ":")
				if len(parts) >= 2 {
					return parts[len(parts)-1] // owner/repo part
				}
			}

			// Handle HTTPS format (https://github.com/owner/repo)
			if strings.Contains(url, "://") {
				parts := strings.Split(url, "/")
				if len(parts) >= 2 {
					// Get last two parts: owner/repo
					owner := parts[len(parts)-2]
					repo := parts[len(parts)-1]
					return owner + "/" + repo
				}
			}

			// Fallback: just get the last part
			parts := strings.Split(url, "/")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
	}

	// Fall back to directory name if git remote fails
	if wd, err := os.Getwd(); err == nil {
		return filepath.Base(wd)
	}

	return ""
}

// getCurrentBranch returns the current git branch
func getCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	branch := strings.TrimSpace(string(output))
	// Handle detached HEAD state
	if branch == "HEAD" {
		return ""
	}

	return branch
}

// IsGitRepo checks if the current directory is a git repository
func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}
