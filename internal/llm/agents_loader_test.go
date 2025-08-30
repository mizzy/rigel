package llm

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadAgentsMD(t *testing.T) {
	t.Run("loads AGENTS.md when it exists", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "test-agents")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Change to the temporary directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		// Create a test AGENTS.md file
		testContent := "# Test AGENTS.md\nThis is test content."
		err = os.WriteFile("AGENTS.md", []byte(testContent), 0644)
		require.NoError(t, err)

		// Load the file
		content, err := LoadAgentsMD()
		assert.NoError(t, err)
		assert.Equal(t, testContent, content)
	})

	t.Run("returns empty string when AGENTS.md doesn't exist", func(t *testing.T) {
		// Create a temporary directory without AGENTS.md
		tmpDir, err := os.MkdirTemp("", "test-no-agents")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Change to the temporary directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		// Load the file (should not exist)
		content, err := LoadAgentsMD()
		assert.NoError(t, err)
		assert.Empty(t, content)
	})
}

func TestPrependAgentsContext(t *testing.T) {
	t.Run("prepends AGENTS.md content when available", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "test-prepend")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Change to the temporary directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		// Create a test AGENTS.md file
		agentsContent := "# Repository Overview\nTest repository content."
		err = os.WriteFile("AGENTS.md", []byte(agentsContent), 0644)
		require.NoError(t, err)

		// Test with existing system prompt
		systemPrompt := "You are a helpful assistant."
		result := PrependAgentsContext(systemPrompt)

		assert.Contains(t, result, "Repository Context from AGENTS.md")
		assert.Contains(t, result, agentsContent)
		assert.Contains(t, result, "System Instructions")
		assert.Contains(t, result, systemPrompt)
	})

	t.Run("returns original prompt when AGENTS.md doesn't exist", func(t *testing.T) {
		// Create a temporary directory without AGENTS.md
		tmpDir, err := os.MkdirTemp("", "test-no-prepend")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Change to the temporary directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		// Test with system prompt
		systemPrompt := "You are a helpful assistant."
		result := PrependAgentsContext(systemPrompt)

		// Should return the original prompt unchanged
		assert.Equal(t, systemPrompt, result)
	})

	t.Run("handles empty system prompt with AGENTS.md", func(t *testing.T) {
		// Create a temporary directory
		tmpDir, err := os.MkdirTemp("", "test-empty-prompt")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Change to the temporary directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		// Create a test AGENTS.md file
		agentsContent := "# Repository Overview\nTest repository content."
		err = os.WriteFile("AGENTS.md", []byte(agentsContent), 0644)
		require.NoError(t, err)

		// Test with empty system prompt
		result := PrependAgentsContext("")

		assert.Contains(t, result, "Repository Context from AGENTS.md")
		assert.Contains(t, result, agentsContent)
		// The system instructions section should be empty
		assert.True(t, strings.HasSuffix(strings.TrimSpace(result), "# System Instructions"))
	})
}
