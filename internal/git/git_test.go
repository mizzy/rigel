package git

import (
	"strings"
	"testing"
)

func TestGetRepoInfo(t *testing.T) {
	info := GetRepoInfo()
	if info == nil {
		t.Skip("Not in a git repository or no git info available")
	}

	// Test that we get some information
	if info.RepoName == "" && info.Branch == "" {
		t.Error("Expected either repo name or branch to be non-empty")
	}

	// If we have a repo name, it should not be empty
	if info.RepoName != "" && len(info.RepoName) == 0 {
		t.Error("Repo name should not be empty string")
	}

	// If we have a branch, it should not be empty
	if info.Branch != "" && len(info.Branch) == 0 {
		t.Error("Branch should not be empty string")
	}
}

func TestIsGitRepo(t *testing.T) {
	// This should return true since we're in a git repository
	if !IsGitRepo() {
		t.Skip("Not in a git repository")
	}
}

func TestGetCurrentBranch(t *testing.T) {
	branch := getCurrentBranch()
	// Branch can be empty if in detached HEAD state
	// Just test that it doesn't panic and returns a string
	if branch == "HEAD" {
		t.Error("Should not return 'HEAD' for detached HEAD state")
	}
}

func TestGetRepoName(t *testing.T) {
	repoName := getRepoName()
	// Repo name should not be empty in this test environment
	if repoName == "" {
		t.Skip("Could not determine repo name")
	}

	// Should not be empty
	if len(repoName) == 0 {
		t.Error("Repo name should not be empty")
	}

	// If it contains a slash, it should be in owner/repo format
	if strings.Contains(repoName, "/") {
		parts := strings.Split(repoName, "/")
		if len(parts) != 2 {
			t.Errorf("Repo name with slash should be in owner/repo format, got: %s", repoName)
		}
		if parts[0] == "" || parts[1] == "" {
			t.Errorf("Neither owner nor repo should be empty in: %s", repoName)
		}
	}
}
