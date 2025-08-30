package llm

import (
	"fmt"
	"os"
	"path/filepath"
)

// LoadAgentsMD loads the AGENTS.md file content from the current working directory
func LoadAgentsMD() (string, error) {
	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Look for AGENTS.md in the current directory
	agentsPath := filepath.Join(cwd, "AGENTS.md")

	// Check if file exists
	if _, err := os.Stat(agentsPath); os.IsNotExist(err) {
		// Return empty string if file doesn't exist (not an error)
		return "", nil
	}

	// Read the file
	content, err := os.ReadFile(agentsPath)
	if err != nil {
		return "", fmt.Errorf("failed to read AGENTS.md: %w", err)
	}

	return string(content), nil
}

// PrependAgentsContext prepends AGENTS.md content to the system prompt if available
func PrependAgentsContext(systemPrompt string) string {
	agentsContent, err := LoadAgentsMD()
	if err != nil {
		// Log error but continue without AGENTS.md content
		// We don't want to fail the entire request
		return systemPrompt
	}

	if agentsContent == "" {
		// No AGENTS.md file found
		return systemPrompt
	}

	// Prepend AGENTS.md content with a separator
	contextPrompt := fmt.Sprintf(`# Repository Context from AGENTS.md

%s

---

# System Instructions

%s`, agentsContent, systemPrompt)

	return contextPrompt
}
